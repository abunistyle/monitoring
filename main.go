package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"monitoring/controllers/bigdata"
	"monitoring/controllers/ecshop"
	"monitoring/controllers/ordersync"
	"monitoring/controllers/web"
	"monitoring/global"
	"monitoring/initialize"
	"monitoring/model/common"
	"net/http"
)

var ctx = context.Background()

func main() {
	param := initialize.InitParam()
	fmt.Println(fmt.Sprintf("root path: %s, config file name: %s, module name: %s, debug: %t", *param.RootPath, *param.ConfigFileName, *param.ModuleName, *param.Debug))

	switch *param.ModuleName {
	case "fdweb":
		webMonitor(param)
	case "web":
		webMonitor(param)
	case "ecshop":
		orderSyncMonitor(param)
		ecshopMonitor(param)
	case "bigdata":
		bigdataMonitor()
	case "order_sync":
		orderSyncMonitor(param)
	default:
		fmt.Println("不支持的moduleName")
	}
}

//网站监控
func webMonitor(param common.Param) {
	var db *gorm.DB
	var projectNames []string
	var platforms []string

	if !initialize.InitDB(*param.ModuleName, &db) {
		return
	}

	if *param.ModuleName == "fdweb" {
		projectNames = []string{"floryday", "airydress"}
		platforms = []string{"PC", "H5", "APP"}
	} else {
		projectNames = []string{"elavee"}
		platforms = []string{"PC", "H5"}
	}
	paySuccessMonitor := web.PaySuccessMonitor{DB: db, RedisClient: global.RedisClient, ProjectNames: projectNames, Platforms: platforms, Debug: *param.Debug}
	paySuccessMonitor.Init()
	if *param.Debug {
		paySuccessMonitor.RunMonitor()
	}

	myCron := cron.New()
	_, _ = myCron.AddFunc("10 * * * *", func() {
		paySuccessMonitor.RunMonitor()
	})
	myCron.Start()

	fmt.Println(fmt.Sprintf("application running on port %d", *param.Port))

	http.Handle("/metrics/web/paySuccess", promhttp.Handler())
	var err = http.ListenAndServe(fmt.Sprintf(":%d", *param.Port), nil)
	if err != nil {
		return
	}
}

//erp监控
func ecshopMonitor(param common.Param) {
	var db *gorm.DB

	if !initialize.InitDB(*param.ModuleName, &db) {
		return
	}
	orderCountMonitor := ecshop.OrderCountMonitor{DB: db, Debug: *param.Debug}
	orderCountMonitor.Init()
	if *param.Debug {
		orderCountMonitor.RunMonitor()
	}

	myCron := cron.New()
	_, _ = myCron.AddFunc("10 * * * *", func() {
		orderCountMonitor.RunMonitor()
	})
	myCron.Start()

	fmt.Println(fmt.Sprintf("application running on port %d", *param.Port))

	http.Handle("/metrics/ecshop/orderCount", promhttp.Handler())
	var err = http.ListenAndServe(fmt.Sprintf(":%d", *param.Port), nil)
	if err != nil {
		return
	}
}

//bigdata监控
func bigdataMonitor() {
	azkabanMontior := bigdata.AzkabanMotior{}

	//每小时10分开始调度
	myCron := cron.New()
	_, _ = myCron.AddFunc("10 * * * *", func() {
		azkabanMontior.Sendnotice()
	})
	myCron.Start()

}

func orderSyncMonitor(param common.Param) {
	compareMonitor := ordersync.CompareMonitor{Debug: *param.Debug}

	if !initialize.InitDB("fdweb", &compareMonitor.FdWebDB) {
		return
	}
	if !initialize.InitDB("web", &compareMonitor.WebDB) {
		return
	}
	if !initialize.InitDB("ecshop", &compareMonitor.ErpDB) {
		return
	}
	compareMonitor.Init()
	if *param.Debug {
		compareMonitor.RunMonitor()
	}
	myCron := cron.New()
	_, _ = myCron.AddFunc("*/30 * * * *", func() {
		compareMonitor.RunMonitor()
	})
	myCron.Start()
}
