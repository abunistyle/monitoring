package ecshop

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"monitoring/model/ecshop/order"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type OrderCountMonitor struct {
	DB                      *gorm.DB
	dataTime                time.Time
	Debug                   bool
	TrySuccessRateGaugeVec  *prometheus.GaugeVec //普罗米修斯需要的数据
	weekBeforeCountRateInfo order.CountRateInfo  //一周以前的数据
	countRateInfo           order.CountRateInfo  //即时数据
}

func (o *OrderCountMonitor) Init() {
	o.TrySuccessRateGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ecshop_order_count",
		Help: "订单工单相关监控",
	},
		[]string{"type_code", "type"},
	)
	o.RefreshMonitorData()
	o.RecordMetrics()
}

// RefreshMonitorData 查数据库获取数据
func (o *OrderCountMonitor) RefreshMonitorData() {
	o.dataTime = time.Now()
	newYorkLocation, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Println("load Asia/Shanghai location failed", err)
		return
	}
	currentTime := time.Unix(o.dataTime.Unix(), 0).In(newYorkLocation)
	{
		m, _ := time.ParseDuration("-168h") //1 week
		endTime := currentTime.Add(m)
		m, _ = time.ParseDuration("-180h") //1 week and 12 hour
		startTime := currentTime.Add(m)
		o.weekBeforeCountRateInfo = o.GetMonitorData(startTime, endTime)
	}
	{
		endTime := currentTime
		m, _ := time.ParseDuration("-12h") //12 hour
		startTime := currentTime.Add(m)
		o.countRateInfo = o.GetMonitorData(startTime, endTime)
	}
	o.countRateInfo.Compare(&o.weekBeforeCountRateInfo)
	return
}

func (o *OrderCountMonitor) GetMonitorData(startTime time.Time, endTime time.Time) order.CountRateInfo {
	var result = order.CountRateInfo{}
	var waitGroup sync.WaitGroup
	for key, value := range o.countRateInfo.GetTypes() {
		waitGroup.Add(1)
		key := key
		value := value
		go func() {
			result.ValueOf(key).TypeCode = key
			result.ValueOf(key).Type = value
			result.ValueOf(key).LastHourCount = o.GetCountData(key, startTime, endTime)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	return result
}
func (o *OrderCountMonitor) GetCountData(name string, startTime time.Time, endTime time.Time) int64 {
	var count int64
	switch name {
	case "orderSync": //网站订单同步
		o.DB.Table("ecshop.ecs_order_info").Where("order_time between ? and ?", startTime, endTime).Count(&count)
	case "orderConfirmed": //订单确认
		o.DB.Table("ecshop.ecs_order_info").Where("confirm_time between ? and ?", startTime.Unix(), endTime.Unix()).Count(&count)
	case "orderReserved": //订单预定
		o.DB.Table("romeo.order_reserve").Where("status = ? and created_at between ? and ?", 1, startTime, endTime).Count(&count)
	case "dispatch": //采购工单
		o.DB.Table("romeo.dispatch_list").Where("CREATED_STAMP between ? and ?", startTime, endTime).Count(&count)
	case "dispatchReceive": //工单收货
		o.DB.Table("mps.pms_dispatch_log").Where("content = ? and created_at between ? and ? ", "收货", startTime, endTime).Count(&count)
	case "dispatchQc": //质检合格
		o.DB.Table("mps.pms_dispatch_log").Where("content = ? and created_at between ? and ? ", "质检合格", startTime, endTime).Count(&count)
	case "orderIn": //下发入库
		o.DB.Table("ecshop.wms_v2_request_log").Where("type = ? and created_at between ? and ?", 20, startTime, endTime).Count(&count)
	case "inCallBack": //入库回调
		o.DB.Table("ecshop.wms_v2_callback_log").Where("type = ? and created_at between ? and ?", 20, startTime, endTime).Count(&count)
	case "orderOut": //下发出库
		o.DB.Table("ecshop.wms_v2_request_log").Where("type = ? and created_at between ? and ?", 30, startTime, endTime).Count(&count)
	case "outCallBack": //出库回调
		o.DB.Table("ecshop.wms_v2_callback_log").Where("type = ? and created_at between ? and ?", 30, startTime, endTime).Count(&count)
	default:
	}
	return count
}

// RunMonitor 创建结构体及对应的指标信息
func (o *OrderCountMonitor) RunMonitor() {
	o.RefreshMonitorData()
	o.RecordMetrics()
	o.SendNotice()
}

func (o *OrderCountMonitor) RecordMetrics() {
	for key, _ := range o.countRateInfo.GetTypes() {
		o.TrySuccessRateGaugeVec.With(prometheus.Labels{
			"type_code": o.countRateInfo.ValueOf(key).TypeCode,
			"type":      o.countRateInfo.ValueOf(key).Type,
		}).Set(float64(o.countRateInfo.ValueOf(key).LastHourCount))
	}
}

func (o *OrderCountMonitor) SendNotice() {
	var messageList []string
	rule := 0.6

	for key, value := range o.countRateInfo.GetTypes() {
		if o.countRateInfo.ValueOf(key).LastHourRateChange < rule {
			rate := strconv.FormatFloat((1-o.countRateInfo.ValueOf(key).LastHourRateChange)*100, 'f', 1, 64)
			message := fmt.Sprintf("%s:近十二小时数据量 %d,比一周前同期下降 %s%%;", value, o.countRateInfo.ValueOf(key).LastHourCount, rate)
			messageList = append(messageList, message)
		}
	}
	if len(messageList) > 0 {
		message := "\n[ERP核心流程数据监控/" + o.dataTime.Format("2006-01-02 15:04:05") + "]\n" + strings.Join(messageList, "\n")
		fmt.Println(message)
		o.RunSendNotice(message)
	}
}

func (o *OrderCountMonitor) RunSendNotice(message string) {
	if o.Debug {
		return
	}
	//if o.Debug {
	//	message = "(测试中，请忽略)" + message
	//}
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://voice.abunistyle.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s", url.QueryEscape(message)))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	}()
}
