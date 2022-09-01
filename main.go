package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"io/ioutil"
	"monitoring/config"
	"monitoring/controllers/ecshop"
	"monitoring/controllers/web"
	"monitoring/global"
	"monitoring/initialize"
	"net/http"
	"os"
)

var ctx = context.Background()
var debug *bool

func main() {
	rootPath := flag.String("r", "./", "请输入程序根路径")
	configFileName := flag.String("c", "./config.yaml", "请输入配置文件路径")
	moduleName := flag.String("m", "web", "请输入模块名，比如：web、fdweb")
	debug = flag.Bool("debug", false, "是否debug，比如：true、false")
	paramPort := flag.Uint64("p", 0, "请输入端口号，比如：8080")
	flag.Parse()
	file, err := ioutil.ReadFile(*configFileName)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(fmt.Sprintf("root path: %s, config file name: %s, module name: %s, debug: %t", *rootPath, *configFileName, *moduleName, *debug))

	//yaml文件内容影射到结构体中
	var myconfig config.Config
	err1 := yaml.Unmarshal(file, &myconfig)
	if err1 != nil {
		fmt.Println(err1)
		os.Exit(1)
		return
	}

	port := (func() uint64 {
		if *paramPort > 0 {
			return *paramPort
		} else {
			return myconfig.Application.Port
		}
	})()
	initialize.InitDBList(&myconfig)
	orderDB, dbExist := global.GetDBByName(*moduleName, "")
	if !dbExist {
		err := errors.New(fmt.Sprintf("module:%s, db not exist!", *moduleName))
		panic(err)
	}
	fmt.Println(fmt.Sprintf("application running on port %d", port))
	switch *moduleName {
	case "fdweb":
		webMonitor(orderDB, moduleName, port)
	case "web":
		webMonitor(orderDB, moduleName, port)
	case "ecshop":
		ecshopMonitor(orderDB, port)
	default:
		fmt.Println("不支持的moduleName")
	}
}

//网站监控
func webMonitor(orderDB *gorm.DB, moduleName *string, port uint64) {
	var projectNames []string
	var platforms []string
	if *moduleName == "fdweb" {
		projectNames = []string{"floryday", "airydress"}
		platforms = []string{"PC", "H5", "APP"}
	} else {
		projectNames = []string{"elavee"}
		platforms = []string{"PC", "H5"}
	}
	paySuccessMonitor := web.PaySuccessMonitor{DB: orderDB, RedisClient: global.RedisClient, ProjectNames: projectNames, Platforms: platforms, Debug: *debug}
	paySuccessMonitor.Init()
	if *debug {
		paySuccessMonitor.RunMonitor()
	}

	myCron := cron.New()
	_, _ = myCron.AddFunc("10 * * * *", func() {
		paySuccessMonitor.RunMonitor()
	})
	myCron.Start()

	http.Handle("/metrics/web/paySuccess", promhttp.Handler())
	var err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		return
	}
}

//erp监控
func ecshopMonitor(orderDB *gorm.DB, port uint64) {
	orderCountMonitor := ecshop.OrderCountMonitor{DB: orderDB, Debug: *debug}
	orderCountMonitor.Init()
	if *debug {
		orderCountMonitor.RunMonitor()
	}

	myCron := cron.New()
	_, _ = myCron.AddFunc("10 * * * *", func() {
		orderCountMonitor.RunMonitor()
	})
	myCron.Start()

	http.Handle("/metrics/ecshop/orderCount", promhttp.Handler())
	var err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		return
	}
}
