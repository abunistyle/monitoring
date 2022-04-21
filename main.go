package main

import (
    "flag"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
    "monitoring/config"
    "monitoring/global"
    "monitoring/initialize"
    "os"
    "reflect"
    "strings"
)

func main() {
    module := flag.String("m", "web", "请输入模块名，比如:web、erp")
    group := flag.String("g", "", "请输入group名，比如:order、goods")
    fun := flag.String("f", "all", "请输入fun名，多个fun用逗号分隔，比如:gmv")

    flag.Parse() // 解析参数
    log.Printf("module:%s, group:%s, fun:%s\n", *module, *group, *fun)

    file, err := ioutil.ReadFile("./config.yaml")
    if err != nil {
        fmt.Print(err)
    }
    var funList []string
    if *fun != "all" {
        funList = strings.Split(*fun, ",")
    }

    //yaml文件内容影射到结构体中
    var myconfig config.Config
    //myconfig := make(map[interface{}]interface{})
    err1 := yaml.Unmarshal(file, &myconfig)
    if err1 != nil {
        fmt.Println(err1)
        os.Exit(1)
        return
    }
    dbConfig, err := initialize.GetMysqlConfig(&myconfig, *module, *group)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
        return
    }
    fmt.Println(dbConfig)
    global.DB = initialize.GormMysqlByConfig(dbConfig)
    if global.DB == nil {
        panic("连接数据库失败, error=dbName is empty")
    }

    router := fmt.Sprintf("%s.%s", *module, *group)
    t := initialize.GetType(router)
    if t == nil {
        panic(fmt.Sprintf("struct %s not exist", router)) // 这里只是演示实际处理根据你的真实使用情况来定
    }
    s := reflect.New(t).Interface()
    method := reflect.ValueOf(s).MethodByName("Run")
    if !method.IsValid() {
        panic("method Run not exist")
    }
    param := []reflect.Value{reflect.ValueOf(funList)}
    method.Call(param)
}
