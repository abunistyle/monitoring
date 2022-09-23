package order


type CouponPriceRateMonitorHistory struct {
    MonitorList []map[string]float64 `json:"monitorInfo"`
    ProjectName string               `json:"projectName"`
}

type CouponPriceRateRule struct {
    Rate float64 `json:"rate"`
}
