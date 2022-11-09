package ordersync

import (
    "fmt"
    "gorm.io/gorm"
    "monitoring/model/ordersync"
    "net/http"
    "net/url"
    "strings"
    "sync"
)

type CompareMonitor struct {
    FdWebDB         *gorm.DB
    WebDB           *gorm.DB
    ErpDB           *gorm.DB
    Debug           bool
    CompareInfoList []ordersync.CompareInfo
}

func (c *CompareMonitor) Init() {
    c.CompareInfoList = append(c.CompareInfoList, ordersync.CompareInfo{Party: "floryday", ProjectName: "floryday", WebDB: c.FdWebDB, ErpDB: c.ErpDB})
    c.CompareInfoList = append(c.CompareInfoList, ordersync.CompareInfo{Party: "airydress", ProjectName: "airydress", WebDB: c.FdWebDB, ErpDB: c.ErpDB})
    c.CompareInfoList = append(c.CompareInfoList, ordersync.CompareInfo{Party: "chicmery", ProjectName: "elavee", WebDB: c.WebDB, ErpDB: c.ErpDB})
}

func (c *CompareMonitor) RunMonitor() {
    c.RefreshMonitorData()
    c.SendNotice()
}

func (c *CompareMonitor) RefreshMonitorData() {
    c.CompareInfoList = c.GetCompareData()
}

func (c *CompareMonitor) GetCompareData() []ordersync.CompareInfo {
    var waitGroup sync.WaitGroup
    var compareInfoList []ordersync.CompareInfo

    for _, compareInfo := range c.CompareInfoList {
        waitGroup.Add(1)
        compareInfo := compareInfo
        go func() {
            compareInfo.Compare()
            compareInfoList = append(compareInfoList, compareInfo)
            waitGroup.Done()
        }()
    }
    waitGroup.Wait()
    return compareInfoList
}

func (c *CompareMonitor) SendNotice() {
    for _, compareInfo := range c.CompareInfoList {
        if len(compareInfo.DiffOrderSnHour) >= 1 {
            message := "\n[网站订单同步ERP异常]\n"
            message = message + "组织：" + compareInfo.Party + "\n"
            message = message + "订单编号：" + strings.Join(compareInfo.DiffOrderSnHour, ",") + "\n"
            message = message + "订单同步时间超过 1 hour"
            fmt.Println(message)
            c.RunSendNotice(message)
        }
        if len(compareInfo.DiffOrderSnThirtyMinutes) >= 10 {
            message := "\n[网站订单同步ERP异常]\n"
            message = message + "组织：" + compareInfo.Party + "\n"
            message = message + "订单编号：" + strings.Join(compareInfo.DiffOrderSnThirtyMinutes, ",") + "\n"
            message = message + "订单同步时间超过 30mins"
            fmt.Println(message)
            c.RunSendNotice(message)
        }
    }
}
func (c *CompareMonitor) RunSendNotice(message string) {
    if c.Debug {
        return
    }
    //if c.Debug {
    //	message = "(测试中，请忽略)" + message
    //}
    go func() {
        resp, err := http.Get(fmt.Sprintf("http://voice.arch800.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s", url.QueryEscape(message)))
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(resp)
    }()
}
