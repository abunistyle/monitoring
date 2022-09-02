package order

type CountRateInfo struct {
	OrderSync       CountData `json:"orderSync"`       //1、网站订单同步
	OrderConfirmed  CountData `json:"orderConfirmed"`  //2、订单确认
	OrderReserved   CountData `json:"orderReserved"`   //3、订单预定
	Dispatch        CountData `json:"dispatch"`        //4、采购工单
	DispatchReceive CountData `json:"dispatchReceive"` //5、工单收货
	DispatchQc      CountData `json:"dispatchQc"`      //6、质检合格
	OrderIn         CountData `json:"orderIn"`         //7、下发入库
	InCallBack      CountData `json:"inCallBack"`      //8、入库回调
	OrderOut        CountData `json:"orderOut"`        //9、下发出库
	OutCallBack     CountData `json:"outCallBack"`     //10、出库回调
}

func (cri *CountRateInfo) Compare(cri2 *CountRateInfo) {
	for key, _ := range cri.GetTypes() {
		cri.ValueOf(key).LastHourCountChange = cri.ValueOf(key).LastHourCount - cri2.ValueOf(key).LastHourCount
		if cri2.ValueOf(key).LastHourCount != 0 {
			cri.ValueOf(key).LastHourRateChange = float64(cri.ValueOf(key).LastHourCount) / float64(cri2.ValueOf(key).LastHourCount)
		} else {
			cri.ValueOf(key).LastHourRateChange = 1
		}
	}
}

func (cri CountRateInfo) GetTypes() map[string]string {
	return map[string]string{
		"orderSync":       "网站订单同步",
		"orderConfirmed":  "订单确认",
		"orderReserved":   "订单预定",
		"dispatch":        "采购工单",
		"dispatchReceive": "工单收货",
		"dispatchQc":      "质检合格",
		"orderIn":         "下发入库",
		"inCallBack":      "入库回调",
		"orderOut":        "下发出库",
		"outCallBack":     "出库回调",
	}
}

func (cri *CountRateInfo) ValueOf(name string) *CountData {
	switch name {
	case "orderSync":
		return &cri.OrderSync
	case "orderConfirmed":
		return &cri.OrderConfirmed
	case "orderReserved":
		return &cri.OrderReserved
	case "dispatch":
		return &cri.Dispatch
	case "dispatchReceive":
		return &cri.DispatchReceive
	case "dispatchQc":
		return &cri.DispatchQc
	case "orderIn":
		return &cri.OrderIn
	case "inCallBack":
		return &cri.InCallBack
	case "orderOut":
		return &cri.OrderOut
	case "outCallBack":
		return &cri.OutCallBack
	default:
		return nil
	}
}

type CountData struct {
	TypeCode            string  `json:"type_code"`           //类型
	Type                string  `json:"type"`                //类型名称
	LastHourCount       int64   `json:"lastHourCount"`       //十二小时内数量
	LastHourCountChange int64   `json:"lastHourCountChange"` //环比变化量
	LastHourRateChange  float64 `json:"lastHourRateChange"`  //环比变化率
}
