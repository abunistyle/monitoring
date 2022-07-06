package order

type PaySuccess struct {
    ProjectName    string  `json:"projectName"`
    PaymentCode    string  `json:"paymentCode"`
    Platform       string  `json:"platform"`
    TrySuccessRate float64 `json:"trySuccessRate"`
    SuccessRate    float64 `json:"successRate"`
    SuccessCount   int64   `json:"successCount"`
    TryCount       int64   `json:"tryCount"`
    AllCount       int64   `json:"allCount"`
}

type PaySuccessMonitor struct {
    PaySuccess
    TrySuccessRateChange float64 `json:"trySuccessRate"`
    SuccessRateChange    float64 `json:"successRate"`
}
