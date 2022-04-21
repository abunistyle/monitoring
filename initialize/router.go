package initialize

import (
    "monitoring/controllers/web"
    "reflect"
)

var (
    routerMap = map[string]reflect.Type{
        "web.order": reflect.TypeOf(web.Order{}),
    }
)

func GetType(router string) reflect.Type {
    return routerMap[router]
}
