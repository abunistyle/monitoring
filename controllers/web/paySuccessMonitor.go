package web

import (
    "encoding/json"
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "gorm.io/gorm"
    "math"
    "monitoring/model/web/order"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "sync"
    "time"
)

type PaySuccessMonitor struct {
    DB                           *gorm.DB
    ProjectNames                 []string
    Platforms                    []string
    Debug                        bool
    UseTestMessage               bool
    Rules                        map[string]order.PaySuccessRule
    SkipPayments                 map[string]bool
    TrySuccessRateGaugeVec       *prometheus.GaugeVec
    SuccessRateGaugeVec          *prometheus.GaugeVec
    TrySuccessRateChangeGaugeVec *prometheus.GaugeVec
    SuccessRateChangeGaugeVec    *prometheus.GaugeVec

    MonitorData []order.PaySuccessMonitor
}

func (p *PaySuccessMonitor) Init() {
    rules := make(map[string]order.PaySuccessRule)
    rules["elavee|PC|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
        //TrySuccessRateChange: 0.9,
        //SuccessRateChange:    0.9,
    }
    rules["elavee|PC|paypal"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
        //TrySuccessRateChange: 0.9,
        //SuccessRateChange:    0.9,
    }
    rules["elavee|PC|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
        //TrySuccessRateChange: 0.7,
        //SuccessRateChange:    0.7,
    }

    rules["elavee|H5|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
        //TrySuccessRateChange: 0.7,
        //SuccessRateChange:    0.7,
    }
    rules["elavee|H5|paypal"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
        //TrySuccessRateChange: 0.9,
        //SuccessRateChange:    0.9,
    }
    rules["elavee|H5|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
        //TrySuccessRateChange: 0.7,
        //SuccessRateChange:    0.7,
    }

    rules["floryday|PC|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
    }
    rules["floryday|PC|paypal"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
    }
    rules["floryday|PC|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }
    rules["floryday|H5|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }
    rules["floryday|H5|paypal"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
    }
    rules["floryday|H5|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }

    rules["airydress|PC|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
    }
    rules["airydress|PC|paypal"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
    }
    rules["airydress|PC|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }
    rules["airydress|H5|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }
    rules["airydress|H5|paypal"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.8,
        SuccessRateChange:       0.8,
    }
    rules["airydress|H5|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }
    p.Rules = rules
    p.SkipPayments = map[string]bool{
        "elavee|H5|checkout#sofort": true,
    }

    p.TrySuccessRateGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_try_success_rate",
        Help: "尝试支付成功率",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    p.SuccessRateGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_success_rate",
        Help: "支付成功率",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    p.TrySuccessRateChangeGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_try_success_rate_change",
        Help: "尝试支付成功率同比占比",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    p.SuccessRateChangeGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_success_rate_change",
        Help: "支付成功率同比占比",
    },
        []string{"project_name", "payment_code", "platform"},
    )
}

func (p *PaySuccessMonitor) IsSkip(projectName string, paymentCode string, platform string) bool {
    key := fmt.Sprintf("%s|%s|%s", projectName, platform, strings.ToLower(paymentCode))
    if _, exist := p.SkipPayments[key]; exist {
        return true
    } else {
        return false
    }
}

func (p *PaySuccessMonitor) GetRule(projectName string, paymentCode string, platform string) (order.PaySuccessRule, bool) {
    key := fmt.Sprintf("%s|%s|%s", projectName, platform, strings.ToLower(paymentCode))
    defaultKey := fmt.Sprintf("%s|%s|other", projectName, platform)
    if rule, exist := p.Rules[key]; exist {
        return rule, true
    } else if rule, exist := p.Rules[defaultKey]; exist {
        return rule, true
    } else {
        var rule order.PaySuccessRule
        return rule, false
    }
}

func (p *PaySuccessMonitor) GetPaymentList(startTime time.Time, endTime time.Time) []order.Payment {
    var result []order.Payment
    startTimeStr := startTime.Format("2006-01-02 15:04:05")
    endTimeStr := endTime.Format("2006-01-02 15:04:05")
    sql := fmt.Sprintf("SELECT\n    lower(oi.project_name) AS project_name,\n    p.payment_id,\n    p.payment_code,\n    CASE\n        WHEN POSITION('api.' IN oi.from_domain) = 0 THEN 'PC'\n        WHEN POSITION('lq-App' IN ua.agent_type) > 0 THEN 'APP'\n        ELSE 'H5'\n\tEND AS platform\nFROM order_info oi\nLEFT JOIN payment p on p.payment_id = oi.payment_id\nLEFT JOIN user_agent ua ON ua.user_agent_id = oi.user_agent_id\nWHERE oi.order_time BETWEEN '%s' AND '%s'\n  AND oi.email NOT LIKE '%@tetx.com'\n  AND oi.email NOT LIKE '%@i9i8.com'\n  AND oi.email NOT LIKE '%@qq.com'\n  AND oi.email NOT LIKE '%@163.com'\n  AND oi.email NOT LIKE '%@jjshouse.com'\n  AND oi.email NOT LIKE '%@jenjenhouse.com'\n  AND oi.email NOT LIKE '%@abunistyle.com'\nGROUP BY oi.project_name,oi.payment_id,platform;", startTimeStr, endTimeStr)
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PaySuccessMonitor) GetStatisticsData(projectName string, paymentId int64, paymentCode string, platform string, startTime time.Time, endTime time.Time) order.PaySuccessRateInfo {
    var result = order.PaySuccessRateInfo{
        TrySuccessRateLastest10:      math.NaN(),
        SuccessRateLastest10:         math.NaN(),
        TrySuccessRateLastest100:     math.NaN(),
        SuccessRateLastest100:        math.NaN(),
        TrySuccessRateLastLastest100: math.NaN(),
        SuccessRateLastLastest100:    math.NaN(),
        TrySuccessRateChange:         math.NaN(),
        SuccessRateChange:            math.NaN(),
    }
    orderList := p.GetOrderData(projectName, paymentId, paymentCode, platform, 0, 200, startTime, endTime)
    if len(orderList) == 0 {
        return result
    }
    var trySuccessRateLastest10 float64 = math.NaN()
    var successRateLastest10 float64 = math.NaN()
    var trySuccessRateLastest100 float64 = math.NaN()
    var successRateLastest100 float64 = math.NaN()
    var trySuccessRateLastLastest100 float64 = math.NaN()
    var successRateLastLastest100 float64 = math.NaN()
    var trySuccessRateChange float64 = math.NaN()
    var successRateChange float64 = math.NaN()
    if len(orderList) >= 10 {
        successCnt, tryCnt, allCnt := p.CalculateSuccessRate(orderList[0:10])
        trySuccessRateLastest10, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(tryCnt)), 64)
        successRateLastest10, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(allCnt)), 64)
    }
    if len(orderList) >= 100 {
        successCnt, tryCnt, allCnt := p.CalculateSuccessRate(orderList[0:100])
        trySuccessRateLastest100, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(tryCnt)), 64)
        successRateLastest100, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(allCnt)), 64)
    }
    if len(orderList) >= 200 {
        successCnt, tryCnt, allCnt := p.CalculateSuccessRate(orderList[100:200])
        trySuccessRateLastLastest100, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(tryCnt)), 64)
        successRateLastLastest100, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(allCnt)), 64)
    }

    if trySuccessRateLastest100 != math.NaN() && trySuccessRateLastLastest100 != math.NaN() {
        trySuccessRateChange, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", trySuccessRateLastest100/trySuccessRateLastLastest100), 64)
    }
    if successRateLastest100 != math.NaN() && successRateLastLastest100 != math.NaN() {
        successRateChange, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", successRateLastest100/successRateLastLastest100), 64)
    }

    //fmt.Println(fmt.Sprintf("%f,%f,%f,%f,%f,%f", trySuccessRateLastest10, successRateLastest10, trySuccessRateLastest100, successRateLastest100, trySuccessRateLastLastest100, successRateLastLastest100))

    result = order.PaySuccessRateInfo{
        TrySuccessRateLastest10:      trySuccessRateLastest10,
        SuccessRateLastest10:         successRateLastest10,
        TrySuccessRateLastest100:     trySuccessRateLastest100,
        SuccessRateLastest100:        successRateLastest100,
        TrySuccessRateLastLastest100: trySuccessRateLastLastest100,
        SuccessRateLastLastest100:    successRateLastLastest100,
        TrySuccessRateChange:         trySuccessRateChange,
        SuccessRateChange:            successRateChange,
    }
    return result
}

func (p *PaySuccessMonitor) CalculateSuccessRate(orderList []order.PaySuccessOrder) (int64, int64, int64) {
    allCnt := len(orderList)
    tryCnt := 0
    successCnt := 0
    for _, row := range orderList {
        if row.TryCnt > 0 {
            tryCnt += 1
        }
        if row.PayStatus == 1 || row.PayStatus == 2 {
            successCnt += 1
        }
    }
    return int64(successCnt), int64(tryCnt), int64(allCnt)
}

func (p *PaySuccessMonitor) GetOrderData(projectName string, paymentId int64, paymentCode string, platform string, offset int64, limit int64, startTime time.Time, endTime time.Time) []order.PaySuccessOrder {
    var result []order.PaySuccessOrder
    startTimeStr := startTime.Format("2006-01-02 15:04:05")
    endTimeStr := endTime.Format("2006-01-02 15:04:05")
    sql := fmt.Sprintf("SELECT\n    t0.*,\n    SUM(IF(t0.pt_payment_code = '%s', 1, 0)) AS try_cnt\nFROM\n\t(\n\tSELECT\n\t    oi.project_name,\n\t    oi.order_id,\n\t\toi.order_sn,\n\t\toi.pay_status,\n\t\toi.payment_id,\n\t\tp.payment_code,\n\t\toi.order_time,\n\t\toi.pay_time,\n\t\tCASE\n\t\t\tWHEN POSITION('api.' IN oi.from_domain) = 0 THEN 'PC'\n\t\t\tWHEN POSITION('lq-App' IN ua.agent_type) > 0 THEN 'APP'\n\t\t\tELSE 'H5'\n\t\tEND AS platform,\n\t\tpt.id as pt_id,\n\t\tpt.payment_code as pt_payment_code\n\tFROM\n\t\torder_info oi\n\tLEFT JOIN payment p on p.payment_id = oi.payment_id\n\tLEFT JOIN paypal_txn pt ON pt.order_sn = oi.order_sn\n\tLEFT JOIN user_agent ua ON ua.user_agent_id = oi.user_agent_id\n\tWHERE\n\t    oi.project_name = '%s'\n\t\tAND oi.order_time BETWEEN '%s' AND '%s'\n\t    AND (oi.payment_id = %d OR pt.payment_code = '%s')\n\t\tAND oi.email NOT LIKE '%%@tetx.com'\n\t\tAND oi.email NOT LIKE '%%@i9i8.com'\n\t\tAND oi.email NOT LIKE '%%@qq.com'\n\t\tAND oi.email NOT LIKE '%%@163.com'\n\t\tAND oi.email NOT LIKE '%%@jjshouse.com'\n\t\tAND oi.email NOT LIKE '%%@jenjenhouse.com'\n\t    AND oi.email NOT LIKE '%%@abunistyle.com'\n\tGROUP BY oi.order_sn,pt.payment_code\n\t) t0\nWHERE t0.platform = '%s'\nGROUP BY t0.order_id\nORDER BY t0.order_id DESC\nLIMIT %d, %d", paymentCode, projectName, startTimeStr, endTimeStr, paymentId, paymentCode, platform, offset, limit)
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PaySuccessMonitor) RefreshMonitorData() {
    p.MonitorData = p.GetMonitorData()
}

func (p *PaySuccessMonitor) GetMonitorData() []order.PaySuccessMonitor {
    tmpTime := time.Now()
    newYorkLocation, err := time.LoadLocation("America/Los_Angeles")
    if err != nil {
        fmt.Println("load America/New_York location failed", err)
        return nil
    }

    currentTime := time.Unix(tmpTime.Unix(), 0).In(newYorkLocation)

    var paymentList []order.Payment
    {
        m, _ := time.ParseDuration("-10m")
        endTime := currentTime.Add(m)
        m, _ = time.ParseDuration("-336h") //2 week
        startTime := endTime.Add(m)
        paymentList = p.GetPaymentList(startTime, endTime)
    }

    m, _ := time.ParseDuration("-10m")
    endTime := currentTime.Add(m)
    m, _ = time.ParseDuration("-2160h") //90 day
    startTime := endTime.Add(m)

    var waitGroup sync.WaitGroup
    var result []order.PaySuccessMonitor
    for _, payment := range paymentList {
        waitGroup.Add(1)
        payment := payment
        go func() {
            statisticsData := p.GetStatisticsData(payment.ProjectName, payment.PaymentId, payment.PaymentCode, payment.Platform, startTime, endTime)
            //fmt.Println(statisticsData)
            resultRow := order.PaySuccessMonitor{
                TrySuccessRateLastest10:      statisticsData.TrySuccessRateLastest10,
                SuccessRateLastest10:         statisticsData.SuccessRateLastest10,
                TrySuccessRateLastest100:     statisticsData.TrySuccessRateLastest100,
                SuccessRateLastest100:        statisticsData.SuccessRateLastest100,
                TrySuccessRateLastLastest100: statisticsData.TrySuccessRateLastLastest100,
                SuccessRateLastLastest100:    statisticsData.SuccessRateLastLastest100,
                TrySuccessRateChange:         statisticsData.TrySuccessRateChange,
                SuccessRateChange:            statisticsData.SuccessRateChange,
                ProjectName:                  payment.ProjectName,
                PaymentCode:                  payment.PaymentCode,
                Platform:                     payment.Platform,
            }
            result = append(result, resultRow)
            waitGroup.Done()
        }()
    }
    waitGroup.Wait()
    //fmt.Println(fmt.Sprintf("result:%d", len(result)))
    return result
}

//创建结构体及对应的指标信息
func (p *PaySuccessMonitor) RunMonitor() {
    p.RefreshMonitorData()
    p.RecordMetrics(p.TrySuccessRateGaugeVec, p.SuccessRateGaugeVec, p.TrySuccessRateChangeGaugeVec, p.SuccessRateChangeGaugeVec)
    p.SendNotice()
}

//创建结构体及对应的指标信息
func (p *PaySuccessMonitor) RecordMetrics(trySuccessRateGaugeVec *prometheus.GaugeVec, successRateGaugeVec *prometheus.GaugeVec, trySuccessRateChangeGaugeVec *prometheus.GaugeVec, successRateChangeGaugeVec *prometheus.GaugeVec) {
    //monitorData := p.GetMonitorData()
    //fmt.Println(monitorData)
    for _, row := range p.MonitorData {
        //bytes1, err := json.Marshal(&row)
        //if err == nil {
        //    fmt.Println("row: ", string(bytes1))
        //}
        if !math.IsNaN(row.TrySuccessRateLastest10) {
            trySuccessRateGaugeVec.With(prometheus.Labels{
                "project_name": row.ProjectName,
                "payment_code": row.PaymentCode,
                "platform":     row.Platform,
            }).Set(row.TrySuccessRateLastest10)
        }
        if !math.IsNaN(row.SuccessRateLastest10) {
            successRateGaugeVec.With(prometheus.Labels{
                "project_name": row.ProjectName,
                "payment_code": row.PaymentCode,
                "platform":     row.Platform,
            }).Set(row.SuccessRateLastest10)
        }
        if !math.IsNaN(row.TrySuccessRateChange) {
            trySuccessRateChangeGaugeVec.With(prometheus.Labels{
                "project_name": row.ProjectName,
                "payment_code": row.PaymentCode,
                "platform":     row.Platform,
            }).Set(row.TrySuccessRateChange)
        }
        if !math.IsNaN(row.SuccessRateChange) {
            successRateChangeGaugeVec.With(prometheus.Labels{
                "project_name": row.ProjectName,
                "payment_code": row.PaymentCode,
                "platform":     row.Platform,
            }).Set(row.SuccessRateChange)
        }
    }
}

func (p *PaySuccessMonitor) SendNotice() {
    //monitorData := p.GetMonitorData()
    //fmt.Println(monitorData)
    var successRateMessageList []string
    var trySuccessRateMessageList []string
    var successRateChangeMessageList []string
    var trySuccessRateChangeMessageList []string
    for _, row := range p.MonitorData {
        //fmt.Println(row)
        if p.IsSkip(row.ProjectName, row.PaymentCode, row.Platform) {
            continue
        }
        rule, exist := p.GetRule(row.ProjectName, row.PaymentCode, row.Platform)
        if !exist {
            continue
        }
        bytes1, err := json.Marshal(&row)
        if err == nil {
            fmt.Println("row: ", string(bytes1))
        }
        if !math.IsNaN(row.TrySuccessRateLastest10) && row.TrySuccessRateLastest10 <= rule.TrySuccessRateLastest10 {
            trySuccessRateMessage := fmt.Sprintf("project:%s,支付方式:%s,平台:%s,近10单尝试支付成功率:%f", row.ProjectName, row.PaymentCode, row.Platform, row.TrySuccessRateLastest10)
            trySuccessRateMessageList = append(trySuccessRateMessageList, trySuccessRateMessage)
        }
        if !math.IsNaN(row.SuccessRateLastest10) && row.SuccessRateLastest10 <= rule.SuccessRateLastest10 {
            successRateMessage := fmt.Sprintf("project:%s,支付方式:%s,平台:%s,近10单支付成功率:%f", row.ProjectName, row.PaymentCode, row.Platform, row.SuccessRateLastest10)
            successRateMessageList = append(successRateMessageList, successRateMessage)
        }
        if !math.IsNaN(row.TrySuccessRateChange) && row.TrySuccessRateChange < rule.TrySuccessRateChange {
            trySuccessRateChangeMessage := fmt.Sprintf("project:%s,支付方式:%s,平台:%s,尝试支付成功率同比<%f:%f", row.ProjectName, row.PaymentCode, row.Platform, rule.TrySuccessRateChange, row.TrySuccessRateChange)
            trySuccessRateChangeMessageList = append(trySuccessRateChangeMessageList, trySuccessRateChangeMessage)
        }
        if !math.IsNaN(row.SuccessRateChange) && row.SuccessRateChange < rule.SuccessRateChange {
            successRateChangeMessage := fmt.Sprintf("project:%s,支付方式:%s,平台:%s,支付成功率同比<%f:%f", row.ProjectName, row.PaymentCode, row.Platform, rule.SuccessRateChange, row.SuccessRateChange)
            successRateChangeMessageList = append(successRateChangeMessageList, successRateChangeMessage)
        }
    }
    if len(trySuccessRateMessageList) > 0 {
        trySuccessRateMessage := "[尝试支付成功率]\n" + strings.Join(trySuccessRateMessageList, "\n")
        fmt.Println(trySuccessRateMessage)
        p.RunSendNotice(trySuccessRateMessage)
    }
    if len(successRateMessageList) > 0 {
        successRateMessage := "[支付成功率]\n" + strings.Join(successRateMessageList, "\n")
        fmt.Println(successRateMessage)
        p.RunSendNotice(successRateMessage)
    }
    if len(trySuccessRateChangeMessageList) > 0 {
        trySuccessRateChangeMessage := "[尝试支付成功率同比变化]\n" + strings.Join(trySuccessRateChangeMessageList, "\n")
        fmt.Println(trySuccessRateChangeMessage)
        p.RunSendNotice(trySuccessRateChangeMessage)
    }
    if len(successRateChangeMessageList) > 0 {
        successRateChangeMessage := "[支付成功率同比变化]\n" + strings.Join(successRateChangeMessageList, "\n")
        fmt.Println(successRateChangeMessage)
        p.RunSendNotice(successRateChangeMessage)
    }
}

func (p *PaySuccessMonitor) RunSendNotice(message string) {
    if p.Debug {
        return
    }
    if p.UseTestMessage {
        message = "(测试中，请忽略)" + message
    }
    go func() {
        resp, err := http.Get(fmt.Sprintf("http://voice.abunistyle.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s", url.QueryEscape(message)))
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(resp)
    }()
}
