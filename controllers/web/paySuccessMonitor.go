package web

import (
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "gorm.io/gorm"
    "monitoring/model/web/order"
    "time"
)

type PaySuccessMonitor struct {
    DB *gorm.DB
}

func (p *PaySuccessMonitor) GetOriginData(startTime time.Time, endTime time.Time) []order.PaySuccess {
    var result []order.PaySuccess
    startTimeStr := startTime.Format("2006-01-02 15:04:05")
    endTimeStr := endTime.Format("2006-01-02 15:04:05")
    sql := fmt.Sprintf("SELECT\n    project_name,\n    payment_code,\n    platform,\n    SUM(IF(pt_id IS NOT NULL AND pay_status IN (1,2), 1, 0)) / SUM(IF(pt_id IS NOT NULL , 1, 0)) AS `try_success_rate`,\n    SUM(IF(pt_id IS NOT NULL AND pay_status IN (1,2), 1, 0)) / COUNT(0) AS `success_rate`,\n    SUM(IF(pt_id IS NOT NULL AND pay_status IN (1,2), 1, 0)) AS `success_count`,\n    SUM(IF(pt_id IS NOT NULL , 1, 0)) AS `try_count`,\n    COUNT(0) as `all_count`\nFROM\n\t(\n\tSELECT\n\t    oi.project_name,\n\t\toi.order_sn,\n\t\toi.pay_status,\n\t\toi.payment_id,\n\t\tp.payment_code,\n\t\toi.order_time,\n\t\toi.pay_time,\n\t\tLEFT (oi.order_time,13) AS order_hour,\n\t\tCASE\n\t\t\tWHEN POSITION('api.' IN oi.from_domain) = 0 THEN 'PC'\n\t\t\tWHEN POSITION('lq-App' IN ua.agent_type) > 0 THEN 'APP'\n\t\t\tELSE 'H5'\n\t\tEND AS platform,\n\t\tpt.id as pt_id,\n\t\tpt.payment_code as pt_payment_code\n\tFROM\n\t\torder_info oi\n\tLEFT JOIN payment p on p.payment_id = oi.payment_id\n\tLEFT JOIN paypal_txn pt ON pt.order_sn = oi.order_sn\n\tLEFT JOIN user_agent ua ON ua.user_agent_id = oi.user_agent_id\n\tWHERE\n\t\toi.order_time BETWEEN '%s' AND '%s'\n\t\tAND oi.email NOT LIKE '%@tetx.com'\n\t\tAND oi.email NOT LIKE '%@i9i8.com'\n\t\tAND oi.email NOT LIKE '%@qq.com'\n\t\tAND oi.email NOT LIKE '%@163.com'\n\t\tAND oi.email NOT LIKE '%@jjshouse.com'\n\t\tAND oi.email NOT LIKE '%@jenjenhouse.com'\n\t    AND oi.email NOT LIKE '%@abunistyle.com'\n\tGROUP BY oi.order_sn,pt.payment_code\n\t) t0\nGROUP BY\n    t0.project_name,\n    t0.payment_id,\n    t0.platform", startTimeStr, endTimeStr)
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PaySuccessMonitor) GetMonitorData() []order.PaySuccessMonitor {
    tmpTime := time.Now()
    newYorkLocation, err := time.LoadLocation("America/Los_Angeles")
    if err != nil {
        fmt.Println("load America/New_York location failed", err)
        return nil
    }
    currentTime := time.Unix(tmpTime.Unix(), 0).In(newYorkLocation)
    m, _ := time.ParseDuration("-10m")
    endTime := currentTime.Add(m)
    m, _ = time.ParseDuration("-3h")
    startTime := endTime.Add(m)

    originData := p.GetOriginData(startTime, endTime)
    if originData == nil {
        return nil
    }

    m, _ = time.ParseDuration("-72h")
    endTime2 := endTime.Add(m)
    startTime2 := startTime.Add(m)

    lastWeekOriginData := p.GetOriginData(startTime2, endTime2)

    originDataMap := make(map[string]order.PaySuccess)
    for _, row := range originData {
        key := fmt.Sprintf("%s|%s|%s", row.ProjectName, row.Platform, row.PaymentCode)
        originDataMap[key] = row
    }
    lastWeekOriginDataMap := make(map[string]order.PaySuccess)
    for _, row := range lastWeekOriginData {
        key := fmt.Sprintf("%s|%s|%s", row.ProjectName, row.Platform, row.PaymentCode)
        lastWeekOriginDataMap[key] = row
    }

    var result []order.PaySuccessMonitor
    for key, row := range originDataMap {
        lastRow, lastRowExist := lastWeekOriginDataMap[key]
        //bytes1, err := json.Marshal(&row)
        //if err == nil {
        //    // 返回的是字节数组 []byte
        //    fmt.Println("row: ", string(bytes1))
        //}
        //bytes2, err2 := json.Marshal(&row)
        //if err2 == nil {
        //    // 返回的是字节数组 []byte
        //    fmt.Println(", last row: ", string(bytes2))
        //}
        var resultRow order.PaySuccessMonitor
        if lastRowExist {
            resultRow = order.PaySuccessMonitor{
                PaySuccess:           row,
                TrySuccessRateChange: (lastRow.TrySuccessRate - row.TrySuccessRate) / lastRow.TrySuccessRate,
                SuccessRateChange:    (lastRow.SuccessRate - row.SuccessRate) / lastRow.SuccessRate,
            }
        } else {
            resultRow = order.PaySuccessMonitor{
                PaySuccess:           row,
                TrySuccessRateChange: 0,
                SuccessRateChange:    0,
            }
        }
        result = append(result, resultRow)
    }

    return result
}

//创建结构体及对应的指标信息
func (p *PaySuccessMonitor) NewPaySuccessMonitor() {
    monitorData := p.GetMonitorData()
    trySuccessRateGaugeVec := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_try_success_rate",
        Help: "尝试支付成功率",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    successRateGaugeVec := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_success_rate",
        Help: "支付成功率",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    trySuccessRateChangeGaugeVec := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_try_success_rate_change",
        Help: "尝试支付成功率同比变化率",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    successRateChangeGaugeVec := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_success_rate_change",
        Help: "支付成功率同比变化率",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    for _, row := range monitorData {
        trySuccessRateGaugeVec.With(prometheus.Labels{
            "project_name": row.ProjectName,
            "payment_code": row.PaymentCode,
            "platform":     row.Platform,
        }).Set(row.TrySuccessRate)
        successRateGaugeVec.With(prometheus.Labels{
            "project_name": row.ProjectName,
            "payment_code": row.PaymentCode,
            "platform":     row.Platform,
        }).Set(row.SuccessRate)
        trySuccessRateChangeGaugeVec.With(prometheus.Labels{
            "project_name": row.ProjectName,
            "payment_code": row.PaymentCode,
            "platform":     row.Platform,
        }).Set(row.TrySuccessRateChange)
        successRateChangeGaugeVec.With(prometheus.Labels{
            "project_name": row.ProjectName,
            "payment_code": row.PaymentCode,
            "platform":     row.Platform,
        }).Set(row.SuccessRateChange)
    }
}
