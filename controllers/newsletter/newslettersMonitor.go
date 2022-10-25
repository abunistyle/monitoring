package newsletter

import (
    "fmt"
    "gorm.io/gorm"
    "log"
    "net/http"
    "net/url"
    "strings"
)

type NewslettersMonitor struct {
    DB                    *gorm.DB
    Debug                 bool
    CurrentMonitorDataMap map[string]Newsletters
    LastMonitorDataMap    map[string]Newsletters
}

type Newsletters struct {
    NlCode      string `json:"nl_code"`
    SendCount   int64  `json:"send_count"`
    ArriveCount int64  `json:"arrive_count"`
}

func (p *NewslettersMonitor) Init() {
    p.LastMonitorDataMap = make(map[string]Newsletters)
    p.CurrentMonitorDataMap = make(map[string]Newsletters)
}

//创建结构体及对应的指标信息
func (p *NewslettersMonitor) RunMonitor() {
    log.Println("send and arrive count monitor start!")
    p.LastMonitorDataMap = p.CurrentMonitorDataMap
    p.CurrentMonitorDataMap = p.GetMonitorData()
    var monitorMsgList []string
    for key, monitorData := range p.CurrentMonitorDataMap {
        if _, exist := p.LastMonitorDataMap[key]; !exist {
            continue
        }
        if monitorData.SendCount == p.LastMonitorDataMap[key].SendCount {
            monitorMsg := fmt.Sprintf("NL系统发送任务异常: [%s]\n前10分钟次任务的发送数量：%d，接口成功数：%d\n目前任务的发送数量：%d，接口成功数：%d\n两次对比没有变化，任务执行异常。", monitorData.NlCode, p.LastMonitorDataMap[key].SendCount, p.LastMonitorDataMap[key].ArriveCount, p.CurrentMonitorDataMap[key].SendCount, p.CurrentMonitorDataMap[key].ArriveCount)
            monitorMsgList = append(monitorMsgList, monitorMsg)
        }
    }
    if len(monitorMsgList) > 0 {
        successRateChangeMessage := strings.Join(monitorMsgList, "\n")
        fmt.Println(successRateChangeMessage)
        p.RunSendNotice(successRateChangeMessage)
    }
    log.Println("send and arrive count monitor end!")
}

func (p *NewslettersMonitor) GetMonitorData() map[string]Newsletters {
    monitorData := make(map[string]Newsletters)
    newSletters := p.GetNewSletters()
    for _, newSletter := range newSletters {
        monitorData[newSletter.NlCode] = Newsletters{
            NlCode:      newSletter.NlCode,
            SendCount:   newSletter.SendCount,
            ArriveCount: newSletter.ArriveCount,
        }
    }
    //monitorData["test"] = Newsletters{
    //    NlCode:      "test",
    //    SendCount:   10,
    //    ArriveCount: 8,
    //}
    return monitorData
}

func (p *NewslettersMonitor) GetNewSletters() []Newsletters {
    var newsletters []Newsletters
    sql := "select nl_code,send_count,arrive_count from newsletters where nl_status = 1 and start_time > '2022-10-01' order by 1 desc limit 10;"
    p.DB.Raw(sql).Scan(&newsletters)
    return newsletters
}

func (p *NewslettersMonitor) RunSendNotice(message string) {
    if p.Debug {
        return
    }
    //if p.Debug {
    //    message = "(测试中，请忽略)" + message
    //}
    go func() {
        resp, err := http.Get(fmt.Sprintf("http://voice.abunistyle.com/notice/singleCallByTts?system=Monitoring&type=voice&errorMsg=%s", url.QueryEscape(message)))
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(resp)
    }()
}
