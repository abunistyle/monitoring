package main

import (
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "monitoring/config"
    "monitoring/initialize"
    "os"
)

func main() {
    file, err := ioutil.ReadFile("./config.yaml")
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
    router := initialize.Routers()
    router.Run()
}
