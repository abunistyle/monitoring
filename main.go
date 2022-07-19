package main

import (
    "flag"
    "fmt"
    "github.com/pkg/errors"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "monitoring/config"
    "monitoring/controllers/web"
    "monitoring/global"
    "monitoring/initialize"
    "net/http"
    "os"
)

func main() {
    rootPath := flag.String("r", "./", "请输入程序根路径")
    configFileName := flag.String("c", "./config.yaml", "请输入配置文件路径")
    moduleName := flag.String("m", "web", "请输入模块名，比如：web、fdweb")
    debug := flag.Bool("debug", false, "是否debug，比如：true、false")
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
    port := flag.Uint64("p", myconfig.Application.Port, "请输入端口号，比如：8080")
    initialize.InitDBList(&myconfig)
    //router := initialize.Routers()
    //router.Run()

    // 创建自定义注册表
    //registry := prometheus.NewRegistry()
    //////注册自定义采集器
    //registry.MustRegister(web.NewPaySuccessMonitor())
    //http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))

    //prom)etheus.MustRegister(web.NewPaySuccessMonitor())
    orderDB, dbExist := global.GetDBByName(*moduleName, "")
    if !dbExist {
        err := errors.New(fmt.Sprintf("module:%s, db not exist!", *moduleName))
        panic(err)
    }
    fmt.Println(myconfig.Application.Port)
    fmt.Println(fmt.Sprintf("application running on port %d", myconfig.Application.Port))
    var projectNames []string
    var platforms []string
    var useTestMessage bool
    if *moduleName == "fdweb" {
        projectNames = []string{"floryday", "airydress"}
        platforms = []string{"PC", "H5", "APP"}
        useTestMessage = true
    } else {
        projectNames = []string{"elavee"}
        platforms = []string{"PC", "H5"}
        useTestMessage = false
    }
    paySuccessMonitor := web.PaySuccessMonitor{DB: orderDB, ProjectNames: projectNames, Platforms: platforms, Debug: *debug, UseTestMessage: useTestMessage}
    paySuccessMonitor.Init()
    paySuccessMonitor.SetMonitor()
    http.Handle("/metrics/web/paySuccess", promhttp.Handler())
    http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
