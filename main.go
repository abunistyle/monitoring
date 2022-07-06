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
    configFileName := flag.String("c", "./config.yaml", "请输入配置文件路径")
    file, err := ioutil.ReadFile(*configFileName)
    if err != nil {
        fmt.Print(err)
    }

    //yaml文件内容影射到结构体中
    var myconfig config.Config
    err1 := yaml.Unmarshal(file, &myconfig)
    if err1 != nil {
        fmt.Println(err1)
        os.Exit(1)
        return
    }
    initialize.InitDBList(&myconfig)
    //router := initialize.Routers()
    //router.Run()

    // 创建自定义注册表
    //registry := prometheus.NewRegistry()
    //////注册自定义采集器
    //registry.MustRegister(web.NewPaySuccessMonitor())
    //http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))

    //prom)etheus.MustRegister(web.NewPaySuccessMonitor())
    orderDB, dbExist := global.GetDBByName("web", "order")
    if !dbExist {
        err := errors.New(fmt.Sprintf("module:%s, group:%s, db not exist!", "web", "order"))
        panic(err)
    }
    paySuccessMonitor := web.PaySuccessMonitor{DB: orderDB}
    paySuccessMonitor.NewPaySuccessMonitor()
    http.Handle("/metrics/web/paySuccess", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
