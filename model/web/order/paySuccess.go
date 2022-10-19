package order

import (
    "time"
)

type Payment struct {
    ProjectName string `json:"projectName"`
    PaymentCode string `json:"paymentCode"`
    PaymentId   int64  `json:"paymentId"`
    Platform    string `json:"platform"`
}

type PaySuccessOrder struct {
    ProjectName   string    `json:"projectName"`
    OrderId       int64     `json:"orderId"`
    OrderSn       string    `json:"orderSn"`
    PayStatus     int64     `json:"payStatus"`
    PaymentId     int64     `json:"paymentId"`
    PaymentCode   string    `json:"payment_code"`
    OrderTime     time.Time `json:"order_time"`
    PayTime       time.Time `json:"payTime"`
    Platform      string    `json:"platform"`
    PtId          int64     `json:"ptId"`
    PtPaymentCode string    `json:"ptPaymentCode"`
    TryCnt        int64     `json:"tryCnt"`
}

type PaySuccessRateInfo struct {
    TrySuccessRateLastest10      float64 `json:"trySuccessRateLastest10"` //最近10单
    SuccessRateLastest10         float64 `json:"successRateLastest10"`
    TrySuccessRateLastest100     float64 `json:"trySuccessRateLastest100"` //最近100单
    SuccessRateLastest100        float64 `json:"successRateLastest100"`
    TrySuccessRateLastLastest100 float64 `json:"trySuccessRateLastLastest100"` //最近101-200单
    SuccessRateLastLastest100    float64 `json:"successRateLastLastest100"`
    TrySuccessRateChange         float64 `json:"trySuccessRate"`
    SuccessRateChange            float64 `json:"successRate"`
}

//type PaySuccess struct {
//    ProjectName    string  `json:"projectName"`
//    PaymentCode    string  `json:"paymentCode"`
//    Platform       string  `json:"platform"`
//    TrySuccessRate float64 `json:"trySuccessRate"`
//    SuccessRate    float64 `json:"successRate"`
//    SuccessCount   int64   `json:"successCount"`
//    TryCount       int64   `json:"tryCount"`
//    AllCount       int64   `json:"allCount"`
//}

type PaySuccessMonitor struct {
    TrySuccessRateLastest10      float64  `json:"trySuccessRateLastest10"`
    SuccessRateLastest10         float64  `json:"successRateLastest10"`
    TrySuccessRateLastest100     float64  `json:"trySuccessRateLastest100"`
    SuccessRateLastest100        float64  `json:"successRateLastest100"`
    TrySuccessRateLastLastest100 float64  `json:"trySuccessRateLastLastest100"`
    SuccessRateLastLastest100    float64  `json:"successRateLastLastest100"`
    TrySuccessRateChange         float64  `json:"trySuccessRateChange"`
    SuccessRateChange            float64  `json:"successRateChange"`
    TryOrderSnListInLastest10    []string `json:"tryOrderSnListInLastest10"`
    OrderSnListLastest10         []string `json:"orderSnListLastest10"`
    ProjectName                  string   `json:"projectName"`
    PaymentCode                  string   `json:"paymentCode"`
    Platform                     string   `json:"platform"`
}

type PaySuccessMonitorHistory struct {
    MonitorList []map[string]float64 `json:"monitorInfo"`
    OrderSnList [][]string           `json:"orderSnList"`
    ProjectName string               `json:"projectName"`
    PaymentCode string               `json:"paymentCode"`
    Platform    string               `json:"platform"`
}

type PaySuccessRule struct {
    TrySuccessRateLastest10 float64 `json:"trySuccessRateLastest10"`
    SuccessRateLastest10    float64 `json:"successRateLastest10"`
    TrySuccessRateChange    float64 `json:"trySuccessRateChange"`
    SuccessRateChange       float64 `json:"successRateChange"`
}

type PaypalTxnAnalysis struct {
    PaypalTxnId     string `json:"paypalTxnId"`
    TxnId           string `json:"txnId"`
    OrderSn         string `json:"orderSn"`
    Action          string `json:"action"`
    ResponseCode    string `json:"responseCode"`
    ResponseMessage string `json:"responseMessage"`
    ResultType      int64  `json:"resultType"`
}
