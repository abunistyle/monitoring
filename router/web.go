package router

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/pkg/errors"
    "monitoring/controllers/web"
    "monitoring/global"
)

type Web struct {
}

func (s *Web) InitWebRouter(Router *gin.RouterGroup) {
    orderGroup := Router.Group("metric/web/order")
    orderDB, dbExist := global.GetDBByName("web", "order")
    if !dbExist {
        err := errors.New(fmt.Sprintf("module:%s, group:%s, db not exist!", "web", "order"))
        panic(err)
    }
    order := web.Order{DB: orderDB}
    {
        orderGroup.GET("gmv", order.Gmv) // 获取角色列表
    }
}
