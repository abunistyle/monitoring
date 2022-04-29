package initialize

import (
    "github.com/gin-gonic/gin"
    "monitoring/router"
)

func Routers() *gin.Engine {
    Router := gin.Default()
    PublicGroup := Router.Group("")
    webRouter := router.Web{}
    webRouter.InitWebRouter(PublicGroup)
    return Router

}
