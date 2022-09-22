package ordersync

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type CompareInfo struct {
	Party                    string //Floryday、Airydress、Chicmery
	ProjectName              string //floryday、airydress、elavee
	WebDB                    *gorm.DB
	ErpDB                    *gorm.DB
	WebOrderSnHour           []string //网站端一小时以上符合的订单ID
	ErpOrderSnHour           []string //ERP端一小时以上符合的订单ID
	DiffOrderSnHour          []string //一小时以上的差集订单ID
	WebOrderSnThirtyMinutes  []string //网站端半小时以上符合的订单ID
	ErpOrderSnThirtyMinutes  []string //ERP端半小时以上符合的订单ID
	DiffOrderSnThirtyMinutes []string //半小时以上的差集订单ID
}

func (c *CompareInfo) Compare() {
	tmpTime := time.Now()
	newYorkLocation, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println("load America/New_York location failed", err)
		return
	}
	currentTime := time.Unix(tmpTime.Unix(), 0).In(newYorkLocation)
	{
		m, _ := time.ParseDuration("-24h")
		startTime := currentTime.Add(m)
		m, _ = time.ParseDuration("-1h")
		endTime := currentTime.Add(m)
		//查询网站近一小时支付成功的订单
		c.WebOrderSnHour = c.GetWebOrderSn(startTime, endTime)
		//查询这些订单，在ERP中已存在的
		c.ErpOrderSnHour = c.GetErpOrderSn(c.WebOrderSnHour)
		c.DiffOrderSnHour = c.diff(c.WebOrderSnHour, c.ErpOrderSnHour)
	}
	{
		m, _ := time.ParseDuration("-24h")
		startTime := currentTime.Add(m)
		m, _ = time.ParseDuration("-30m")
		endTime := currentTime.Add(m)
		c.WebOrderSnThirtyMinutes = c.GetWebOrderSn(startTime, endTime)
		c.ErpOrderSnThirtyMinutes = c.GetErpOrderSn(c.WebOrderSnThirtyMinutes)
		c.DiffOrderSnThirtyMinutes = c.diff(c.WebOrderSnThirtyMinutes, c.ErpOrderSnThirtyMinutes)
	}
}

func (c *CompareInfo) GetWebOrderSn(startTime time.Time, endTime time.Time) []string {
	result := make([]string, 0)
	c.WebDB.Table("vbridal.order_info").Where("project_name = ? and order_status != 2 and pay_status = 2 and pay_time >= ? and pay_time <= ?", c.ProjectName, startTime, endTime).Pluck("order_sn", &result)
	return result
}

func (c *CompareInfo) GetErpOrderSn(orderSn []string) []string {
	result := make([]string, 0)
	c.ErpDB.Table("ecshop.ecs_order_info").Where("taobao_order_sn in (?)", orderSn).Pluck("taobao_order_sn", &result)
	return result
}

func (c *CompareInfo) diff(diff1 []string, diff2 []string) []string {
	result := make([]string, 0)
	for _, string1 := range diff1 {
		has := false
		for _, string2 := range diff2 {
			if string1 == string2 {
				has = true
			}
		}
		if !has {
			result = append(result, string1)
		}
	}
	return result
}
