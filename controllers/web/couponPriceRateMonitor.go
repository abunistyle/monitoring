package web

import (
    "fmt"
    "gorm.io/gorm"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "log"
    "math"
    "monitoring/model/web/order"
    "net/http"
    "net/url"
    "time"
)

type CouponPriceRateMonitor struct {
    DB           *gorm.DB
    ProjectNames []string
    Debug        bool
    Rules        map[string]order.CouponPriceRateRule
    RateGaugeVec       *prometheus.GaugeVec

    MonitorDataHistoryMap map[string]*order.CouponPriceRateMonitorHistory
}

type OrderBouncePriceRate struct {
    ProjectName  string     `json:"project_name"`
    Rate         int64      `json:"rate"`
}

type OrderBouncePriceThreeHourRate struct {
    ProjectName  string     `json:"project_name"`
    Rate         float64      `json:"rate"`
    LastHourRate float64      `json:"lastHourRate"`
    LastTwoHourRate float64   `json:"lastTwoHourRate"`
}

func (p *CouponPriceRateMonitor) RunMonitor() {
    log.Println("coupon price rate monitor start!")
    var thresholdHigh float64
    var thresholdLow float64
    p.RateGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "coupon_price_rate",
    },
        []string{"project_name"},
    )
    for _, projectName := range p.ProjectNames {
        rate := p.getCouponPriceRate(projectName)
        if !math.IsNaN(rate.Rate) {
            p.RateGaugeVec.with(prometheus.labels{
                "project_name": projectName,
            }).set(rate.Rate)
        }
        thresholdHigh = 0.3
        thresholdLow = 0.15
        if projectName == "airydress" {
            thresholdHigh = 0.2
            thresholdLow = 0.1
        }
        if rate.Rate >= thresholdHigh && rate.LastHourRate >= thresholdHigh {
            // 连续2小时高于高位 触发高级报警
            msg := fmt.Sprintf("[红包金额占比异常]组织:%s,最近2小时占比分别为:%s,%s. 大于设定阈值:%s,请检查红包设置",projectName,rate.LastHourRate,rate.Rate, thresholdHigh)
            log.Print(msg)
            p.RunSendNotice(msg, false)
        } else if rate.Rate >= thresholdLow && rate.LastHourRate >= thresholdLow && rate.LastTwoHourRate >= thresholdLow {
            // 连续3小时高于中位 触发中级报警
            msg := fmt.Sprintf("[红包金额占比异常]组织:%s,最近3小时占比分别为:%s,%s,%s. 大于设定阈值:%s,请检查红包设置",projectName,rate.LastTwoHourRate,rate.LastHourRate,rate.Rate, thresholdLow)
            log.Print(msg)
            p.RunSendNotice(msg, false)
        }
    }
    log.Println("coupon price rate monitor end!")
}

func (p *CouponPriceRateMonitor) Init() {
    p.MonitorDataHistoryMap = make(map[string]*order.CouponPriceRateMonitorHistory)
    rules := make(map[string]order.CouponPriceRateRule)
    for _, project := range p.ProjectNames {
        thresholdHigh := 0.3
        thresholdLow := 0.15
        if project == "airydress" {
            thresholdHigh = 0.2
            thresholdLow = 0.1
        }
        rules[project + "|high"] = order.CouponPriceRateRule{
            Rate: thresholdHigh,
        }
        rules[project + "|low"] = order.CouponPriceRateRule{
            Rate: thresholdLow,
        }
    }
    p.Rules = rules
}

func (p *CouponPriceRateMonitor) getCouponPriceRate(projectName string) OrderBouncePriceThreeHourRate {
    tmpTime := time.Now()
    newYorkLocation, err := time.LoadLocation("America/Los_Angeles")
    var res OrderBouncePriceThreeHourRate
    if err != nil {
        fmt.Println("load America/New_York location failed", err)
        return res
    }

    currentTime := time.Unix(tmpTime.Unix(), 0).In(newYorkLocation)
    // 前一小时
    m, _ := time.ParseDuration("-0s")
    endTime := currentTime.Add(m)
    m, _ = time.ParseDuration("-1h")
    startTime := endTime.Add(m)
    endTimeStr := endTime.Format("2006-01-02 15:04:05")
    startTimeStr := startTime.Format("2006-01-02 15:04:05")
    rateOne := p.getCouponPriceOneHourRate(projectName, startTimeStr, endTimeStr)
    // 前2小时
    ssTime := startTime.Add(m)
    ssTimeStr := ssTime.Format("2006-01-02 15:04:05")
    rateLastHour := p.getCouponPriceOneHourRate(projectName, ssTimeStr, startTimeStr)
    // 前3小时
    sssTime := startTime.Add(m)
    sssTimeStr := sssTime.Format("2006-01-02 15:04:05")
    rateLastTwoHour := p.getCouponPriceOneHourRate(projectName, sssTimeStr, ssTimeStr)
    res = OrderBouncePriceThreeHourRate{
        ProjectName:     projectName,
        Rate:            float64(rateOne.Rate),
        LastHourRate:    float64(rateLastHour.Rate),
        LastTwoHourRate: float64(rateLastTwoHour.Rate),
    }
    return res
}

func (p *CouponPriceRateMonitor) getCouponPriceOneHourRate(projectName string, startTimeStr string, endTimeStr string) OrderBouncePriceRate {
    var resultList []OrderBouncePriceRate
    var result OrderBouncePriceRate
    sql := fmt.Sprintf(`
SELECT
	oi.project_name, (SUM(oi.integral - oi.bonus) / SUM(oi.goods_amount)) AS 'rate'
FROM order_info oi 
WHERE oi.email NOT REGEXP '@qq.com|@tetx.com|@i9i8.com|@aubnistyle.com|alanyuanzhou@gmail.com'
AND oi.bonus < 0
AND oi.project_name = '%s'
AND oi.order_time >= '%s' AND oi.order_time < '%s';`, projectName, startTimeStr, endTimeStr)
    p.DB.Raw(sql).Scan(&resultList)
    for _, row := range resultList {
        return OrderBouncePriceRate{
            ProjectName: row.ProjectName,
            Rate: row.Rate,
        }
    }
    return result
}


func (p *CouponPriceRateMonitor) RunSendNotice(message string, voice bool) {
    if p.Debug {
        return
    }
    //if p.Debug {
    //    message = "(测试中，请忽略)" + message
    //}
    var voiceStr string
    voiceStr = ""
    if voice {
        voiceStr = "&type=voice"
    }
    go func() {
        resp, err := http.Get(fmt.Sprintf("http://voice.abunistyle.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s%s", url.QueryEscape(message), voiceStr))
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(resp)
    }()
}
