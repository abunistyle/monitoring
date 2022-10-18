package web

import (
    "encoding/json"
    "fmt"
    "github.com/go-redis/redis"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "gorm.io/gorm"
    "log"
    "math"
    "monitoring/model/web/order"
    "monitoring/utils"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "sync"
    "time"
)

const MaxHistoryListLen = 10

const PaypalTxnResultTypeUnknown = 0              //未知
const PaypalTxnResultTypeSuccess = 1              //支付成功
const PaypalTxnResultTypeUserResultType = 2       //用户原因导致的失败
const PaypalTxnResultTypePayCompanyResultType = 3 //支付公司原因导致的失败
const PaypalTxnResultTypeOwnResultType = 4        //程序自身原因导致的失败

type PaySuccessMonitor struct {
    DB                           *gorm.DB
    RedisClient                  *redis.Client
    ProjectNames                 []string
    Platforms                    []string
    Debug                        bool
    Rules                        map[string]order.PaySuccessRule
    SkipPayments                 map[string]bool
    TrySuccessRateGaugeVec       *prometheus.GaugeVec
    SuccessRateGaugeVec          *prometheus.GaugeVec
    TrySuccessRateChangeGaugeVec *prometheus.GaugeVec
    SuccessRateChangeGaugeVec    *prometheus.GaugeVec

    MonitorData            []order.PaySuccessMonitor
    MonitorDataHistroryMap map[string]*order.PaySuccessMonitorHistory
}

func (p *PaySuccessMonitor) Init() {
    p.MonitorDataHistroryMap = make(map[string]*order.PaySuccessMonitorHistory)
    //for _, projectName := range p.ProjectNames {
    //    cacheKey := "monitoring_" + projectName
    //    hashValues := p.RedisClient.HGetAll(cacheKey)
    //    for field, value := range hashValues.Val() {
    //        var historyData order.PaySuccessMonitorHistory
    //        json.Unmarshal([]byte(value), &historyData)
    //        p.MonitorDataHistroryMap[field] = &historyData
    //    }
    //}

    rules := make(map[string]order.PaySuccessRule)
    rules["elavee|PC|checkout"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.75,
        SuccessRateChange:       0.75,
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
        TrySuccessRateChange:    0.7,
        SuccessRateChange:       0.7,
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
        TrySuccessRateChange:    0.75,
        SuccessRateChange:       0.75,
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
        TrySuccessRateChange:    0.7,
        SuccessRateChange:       0.7,
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
        TrySuccessRateChange:    0.65,
        SuccessRateChange:       0.65,
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
        TrySuccessRateChange:    0.7,
        SuccessRateChange:       0.7,
    }
    rules["airydress|H5|other"] = order.PaySuccessRule{
        TrySuccessRateLastest10: 0,
        SuccessRateLastest10:    0,
        TrySuccessRateChange:    0.6,
        SuccessRateChange:       0.6,
    }
    p.Rules = rules
    p.SkipPayments = map[string]bool{
        //支付方式需要小写
        //"elavee|H5|checkout#sofort":         true,
        //"elavee|H5|checkout#giropay":        true,
        "floryday|H5|wire_transfer_vbridal": true,
        "floryday|H5|dlocal#pse":            true,
        //"floryday|H5|checkout#sofort":       true,
        //"floryday|H5|checkout#giropay":      true,
        "floryday|H5|ebanx#oxxo": true,
        "floryday|H5|ebanx":      true,
        "floryday|PC|dlocal#pse": true,
        //"floryday|PC|checkout#sofort":       true,
        "airydress|H5|dlocal":               true,
        "airydress|H5|braintree#creditcard": true,
        "airydress|H5|ebanx":                true,
        //"airydress|H5|checkout":             true,
        "airydress|H5|dlocal#pse": true,
        //"airydress|H5|checkout#sofort":  true,
        "airydress|H5|checkout#giropay": true,
        "airydress|H5|ebanx#oxxo":       true,
        "airydress|PC|dlocal#pse":       true,
        //"airydress|PC|checkout#sofort":  true,
        //"airydress|PC|checkout":             true,
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
        Help: "尝试支付成功率环比占比",
    },
        []string{"project_name", "payment_code", "platform"},
    )
    p.SuccessRateChangeGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "pay_success_success_rate_change",
        Help: "支付成功率环比占比",
    },
        []string{"project_name", "payment_code", "platform"},
    )

    p.RefreshMonitorData()
    p.RecordMetrics(p.TrySuccessRateGaugeVec, p.SuccessRateGaugeVec, p.TrySuccessRateChangeGaugeVec, p.SuccessRateChangeGaugeVec)
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
    sql := fmt.Sprintf(`
SELECT
    lower(oi.project_name) AS project_name,
    p.payment_id,
    p.payment_code,
    CASE
        WHEN POSITION('api.' IN oi.from_domain) = 0 THEN 'PC'
        WHEN POSITION('lq-App' IN ua.agent_type) > 0 THEN 'APP'
        ELSE 'H5'
	END AS platform
FROM order_info oi
LEFT JOIN payment p on p.payment_id = oi.payment_id
LEFT JOIN user_agent ua ON ua.user_agent_id = oi.user_agent_id
WHERE oi.order_time BETWEEN '%s' AND '%s'
  AND oi.payment_id NOT IN (178,179,181,201,203,236,237,238,240)
--   AND p.payment_code = 'CHECKOUT#SOFORT'
  AND oi.email NOT LIKE '%%@tetx.com'
  AND oi.email NOT LIKE '%%@i9i8.com'
  AND oi.email NOT LIKE '%%@qq.com'
  AND oi.email NOT LIKE '%%@163.com'
  AND oi.email NOT LIKE '%%@jjshouse.com'
  AND oi.email NOT LIKE '%%@jenjenhouse.com'
  AND oi.email NOT LIKE '%%@abunistyle.com'
GROUP BY oi.project_name,oi.payment_id,platform;`, startTimeStr, endTimeStr)
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PaySuccessMonitor) GetStatisticsData(projectName string, paymentId int64, paymentCode string, platform string, startTime time.Time, endTime time.Time) (order.PaySuccessRateInfo, []string) {
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
        return result, nil
    }
    var trySuccessRateLastest10 float64 = math.NaN()
    var successRateLastest10 float64 = math.NaN()
    var trySuccessRateLastest100 float64 = math.NaN()
    var successRateLastest100 float64 = math.NaN()
    var trySuccessRateLastLastest100 float64 = math.NaN()
    var successRateLastLastest100 float64 = math.NaN()
    var trySuccessRateChange float64 = math.NaN()
    var successRateChange float64 = math.NaN()
    var orderSnListLastest10 []string
    if len(orderList) >= 10 {
        successCnt, tryCnt, allCnt := p.CalculateSuccessRate(orderList[0:10])
        trySuccessRateLastest10, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(tryCnt)), 64)
        successRateLastest10, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(successCnt)/float64(allCnt)), 64)
        for i := 0; i < 10; i++ {
            orderSnListLastest10 = append(orderSnListLastest10, orderList[i].OrderSn)
        }
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
    return result, orderSnListLastest10
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
    sql := fmt.Sprintf(`
SELECT
    t0.*,
    SUM(IF(t0.pt_payment_code = '%s', 1, 0)) AS try_cnt
FROM
	(
	SELECT
	    oi.project_name,
	    oi.order_id,
		oi.order_sn,
		oi.pay_status,
		oi.payment_id,
		p.payment_code,
		oi.order_time,
		oi.pay_time,
		CASE
			WHEN POSITION('api.' IN oi.from_domain) = 0 THEN 'PC'
			WHEN POSITION('lq-App' IN ua.agent_type) > 0 THEN 'APP'
			ELSE 'H5'
		END AS platform,
		pt.id as pt_id,
		pt.payment_code as pt_payment_code
	FROM
		order_info oi
	LEFT JOIN payment p on p.payment_id = oi.payment_id
	LEFT JOIN paypal_txn pt ON pt.order_sn = oi.order_sn
	LEFT JOIN user_agent ua ON ua.user_agent_id = oi.user_agent_id
	WHERE
	    oi.project_name = '%s'
		AND oi.order_time BETWEEN '%s' AND '%s'
	    AND (oi.payment_id = %d OR pt.payment_code = '%s')
		AND oi.email NOT LIKE '%%@tetx.com'
		AND oi.email NOT LIKE '%%@i9i8.com'
		AND oi.email NOT LIKE '%%@qq.com'
		AND oi.email NOT LIKE '%%@163.com'
		AND oi.email NOT LIKE '%%@jjshouse.com'
		AND oi.email NOT LIKE '%%@jenjenhouse.com'
	    AND oi.email NOT LIKE '%%@abunistyle.com'
	GROUP BY oi.order_sn,pt.payment_code
	) t0
WHERE t0.platform = '%s'
GROUP BY t0.order_id
ORDER BY t0.order_id DESC
LIMIT %d, %d`, paymentCode, projectName, startTimeStr, endTimeStr, paymentId, paymentCode, platform, offset, limit)
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PaySuccessMonitor) RefreshMonitorData() {
    p.MonitorData = p.GetMonitorData()
}

func (p *PaySuccessMonitor) RefreshMonitorDataHistory() {
    for _, row := range p.MonitorData {
        key := fmt.Sprintf("%s|%s|%s", row.ProjectName, row.Platform, strings.ToLower(row.PaymentCode))

        if _, exist := p.MonitorDataHistroryMap[key]; !exist {
            p.MonitorDataHistroryMap[key] = &order.PaySuccessMonitorHistory{
                ProjectName: row.ProjectName,
                Platform:    row.Platform,
                PaymentCode: row.PaymentCode,
            }
        }
        if len(p.MonitorDataHistroryMap[key].MonitorList) >= MaxHistoryListLen {
            p.MonitorDataHistroryMap[key].MonitorList = p.MonitorDataHistroryMap[key].MonitorList[1:]
        }
        if len(p.MonitorDataHistroryMap[key].OrderSnList) >= MaxHistoryListLen {
            p.MonitorDataHistroryMap[key].OrderSnList = p.MonitorDataHistroryMap[key].OrderSnList[1:]
        }

        p.MonitorDataHistroryMap[key].MonitorList = append(p.MonitorDataHistroryMap[key].MonitorList, map[string]float64{
            "trySuccessRateLastest10": row.TrySuccessRateLastest10,
            "successRateLastest10":    row.SuccessRateLastest10,
            "trySuccessRateChange":    row.TrySuccessRateChange,
            "successRateChange":       row.SuccessRateChange,
        })
        p.MonitorDataHistroryMap[key].OrderSnList = append(p.MonitorDataHistroryMap[key].OrderSnList, row.OrderSnListLastest10)
    }
    //for key, row := range p.MonitorDataHistroryMap {
    //    cacheKey := "monitoring_" + row.ProjectName
    //    rowJson, err0 := json.Marshal(&row)
    //    if err0 != nil {
    //        fmt.Println(key, cacheKey)
    //        fmt.Println("hahaha", rowJson, err0)
    //    }
    //    err := p.RedisClient.HSet(cacheKey, key, rowJson).Err()
    //    if err != nil {
    //       fmt.Println(err.Error())
    //       return
    //    }
    //}
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
            statisticsData, orderSnListLastest10 := p.GetStatisticsData(payment.ProjectName, payment.PaymentId, payment.PaymentCode, payment.Platform, startTime, endTime)
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
                OrderSnListLastest10:         orderSnListLastest10,
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
    p.RefreshMonitorDataHistory()
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
        if !math.IsNaN(row.TrySuccessRateLastest10) && row.TrySuccessRateLastest10 <= rule.TrySuccessRateLastest10 && !p.IsIgnoreSendNotice(row.ProjectName, row.PaymentCode, row.Platform, "trySuccessRateLastest10", row.TrySuccessRateLastest10, row.OrderSnListLastest10) && !p.IsIgnoreWithFailureReason(row.PaymentCode, row.OrderSnListLastest10) {
            trySuccessRateMessage := fmt.Sprintf("组织：%s\n支付方式：%s\n平台：%s\n近10单尝试支付成功率:%f", row.ProjectName, row.PaymentCode, row.Platform, row.TrySuccessRateLastest10)
            trySuccessRateMessageList = append(trySuccessRateMessageList, trySuccessRateMessage)
        }
        if !math.IsNaN(row.SuccessRateLastest10) && row.SuccessRateLastest10 <= rule.SuccessRateLastest10 && !p.IsIgnoreSendNotice(row.ProjectName, row.PaymentCode, row.Platform, "successRateLastest10", row.SuccessRateLastest10, row.OrderSnListLastest10) && !p.IsIgnoreWithFailureReason(row.PaymentCode, row.OrderSnListLastest10) {
            successRateMessage := fmt.Sprintf("组织：%s\n支付方式：%s\n平台：%s\n近10单支付成功率:%f", row.ProjectName, row.PaymentCode, row.Platform, row.SuccessRateLastest10)
            successRateMessageList = append(successRateMessageList, successRateMessage)
        }
        if !math.IsNaN(row.TrySuccessRateChange) && row.TrySuccessRateChange < rule.TrySuccessRateChange && !p.IsIgnoreSendNotice(row.ProjectName, row.PaymentCode, row.Platform, "trySuccessRateChange", row.TrySuccessRateChange, row.OrderSnListLastest10) {
            trySuccessRateChangeMessage := fmt.Sprintf("组织：%s\n支付方式：%s\n平台：%s\n近100单尝试支付成功率：%f\n尝试支付成功率环比<%f：%f", row.ProjectName, row.PaymentCode, row.Platform, row.TrySuccessRateLastest100, rule.TrySuccessRateChange, row.TrySuccessRateChange)
            trySuccessRateChangeMessageList = append(trySuccessRateChangeMessageList, trySuccessRateChangeMessage)
        }
        if !math.IsNaN(row.SuccessRateChange) && row.SuccessRateChange < rule.SuccessRateChange && !p.IsIgnoreSendNotice(row.ProjectName, row.PaymentCode, row.Platform, "successRateChange", row.SuccessRateChange, row.OrderSnListLastest10) {
            successRateChangeMessage := fmt.Sprintf("组织：%s\n支付方式：%s\n平台：%s\n近100单支付成功率：%f\n支付成功率环比<%f：%f", row.ProjectName, row.PaymentCode, row.Platform, row.SuccessRateLastest100, rule.SuccessRateChange, row.SuccessRateChange)
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
        trySuccessRateChangeMessage := "[尝试支付成功率环比变化]\n" + strings.Join(trySuccessRateChangeMessageList, "\n")
        fmt.Println(trySuccessRateChangeMessage)
        p.RunSendNotice(trySuccessRateChangeMessage)
    }
    if len(successRateChangeMessageList) > 0 {
        successRateChangeMessage := "[支付成功率环比变化]\n" + strings.Join(successRateChangeMessageList, "\n")
        fmt.Println(successRateChangeMessage)
        p.RunSendNotice(successRateChangeMessage)
    }
}

// IsIgnoreSendNotice 同样的指标最多报三次/**
func (p *PaySuccessMonitor) IsIgnoreSendNotice(projectName string, paymentCode string, platform string, fieldName string, fieldValue float64, orderSnList []string) bool {
    key := fmt.Sprintf("%s|%s|%s", projectName, platform, strings.ToLower(paymentCode))
    if _, exist := p.MonitorDataHistroryMap[key]; !exist {
        return false
    }
    sameCount := 0
    for i := 1; i < 4; i++ {
        index := len(p.MonitorDataHistroryMap[key].MonitorList) - i - 1
        //防止越界
        if index < 0 {
            log.Printf("project:%s,支付方式:%s,平台:%s,%s:%f, 数组越界，报警次数已达%d次，执行本次报警\n", projectName, paymentCode, platform, fieldName, fieldValue, sameCount)
            return false
        }
        if p.MonitorDataHistroryMap[key].OrderSnList[index] == nil || p.MonitorDataHistroryMap[key].MonitorList[index] == nil || math.IsNaN(p.MonitorDataHistroryMap[key].MonitorList[index][fieldName]) {
            log.Printf("project:%s,支付方式:%s,平台:%s,%s:%f, 数据为nil，报警次数已达%d次，执行本次报警\n", projectName, paymentCode, platform, fieldName, fieldValue, sameCount)
            return false
        }
        //判断相等
        if math.Abs(p.MonitorDataHistroryMap[key].MonitorList[index][fieldName]-fieldValue) >= 0.000001 {
            log.Printf("project:%s,支付方式:%s,平台:%s,%s:%f, 数据相等，报警次数已达%d次，执行本次报警\n", projectName, paymentCode, platform, fieldName, fieldValue, sameCount)
            return false
        }

        md5a := utils.MD5(strings.Join(orderSnList, ","))
        md5b := utils.MD5(strings.Join(p.MonitorDataHistroryMap[key].OrderSnList[index], ","))
        if md5a != md5b {
            log.Printf("project:%s,支付方式:%s,平台:%s,%s:%f, 数据有变化，报警次数已达%d次，执行本次报警\n", projectName, paymentCode, platform, fieldName, fieldValue, sameCount)
            return false
        }
        sameCount += 1
    }
    if sameCount >= 3 {
        log.Printf("project:%s,支付方式:%s,平台:%s,%s:%f, 报警次数已达3次，忽略本次报警\n", projectName, paymentCode, platform, fieldName, fieldValue)
        return true
    }
    log.Printf("project:%s,支付方式:%s,平台:%s,%s:%f, 报警次数已达%d次，执行本次报警\n", projectName, paymentCode, platform, fieldName, fieldValue, sameCount)
    return false
}

// IsIgnoreWithFailureReason 分析失败原因，若是用户原因导致的失败，则返回true/**
func (p *PaySuccessMonitor) IsIgnoreWithFailureReason(paymentCode string, orderSnList []string) bool {
    //若是近10个订单都是程序原因导致失败，则发送报警
    for _, orderSn := range orderSnList {
        isOwnReasonFailure := p.IsOwnReasonFailure(paymentCode, orderSn)
        if !isOwnReasonFailure {
            log.Println(fmt.Sprintf("%s,%s, failure not own reason, ignore notice!", orderSn, paymentCode))
            return true
        }
    }
    return false
}

func (p *PaySuccessMonitor) IsOwnReasonFailure(paymentCode string, orderSn string) bool {
    analysisList := p.GetPaypalTxnAnalysis(paymentCode, orderSn)
    actions := []string{"redirect", "pay", "capture"}
    failureResultTypes := []int64{PaypalTxnResultTypeUnknown, PaypalTxnResultTypePayCompanyResultType, PaypalTxnResultTypeOwnResultType}
    for _, analysis := range analysisList {
        if utils.ArrayContains(actions, analysis.Action) && utils.ArrayContains(failureResultTypes, analysis.ResultType) {
            return true
        }
    }
    return false
}

func (p *PaySuccessMonitor) GetPaypalTxnAnalysis(paymentCode string, orderSn string) []order.PaypalTxnAnalysis {
    var result []order.PaypalTxnAnalysis
    sql := fmt.Sprintf(`
select pt.order_sn, pta.*
from paypal_txn pt
left join paypal_txn_analysis pta on pt.id = pta.paypal_txn_id
where pt.order_sn = '%s'
and pt.payment_code = '%s'`, orderSn, paymentCode)
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PaySuccessMonitor) RunSendNotice(message string) {
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
