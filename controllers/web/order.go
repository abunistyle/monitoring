package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"monitoring/model/common"
	"monitoring/model/web/order"
	"monitoring/utils"
	"reflect"
	"sync/atomic"
	"time"
)

type Order struct {
	DB *gorm.DB
}

var (
	RunFunList = []string{"Gmv", "TestGmv"}
)

func (o *Order) Run(funList []string) {
	var myFunList []string
	if funList != nil && len(funList) > 0 {
		for _, fun := range funList {
			if utils.ArrayContains(RunFunList, fun) {
				myFunList = append(myFunList, fun)
			}
		}
	} else {
		myFunList = RunFunList
	}
	fmt.Println(myFunList)

	remainCount := int64(len(myFunList))
	v := reflect.ValueOf(o)
	for _, fun := range myFunList {
		method := v.MethodByName(fun)
		if !method.IsValid() {
			panic(fmt.Sprintf("method %s not exist", fun))
		}
		param := []reflect.Value{reflect.ValueOf(&remainCount)}
		go method.Call(param)
	}

	for {
		if remainCount <= 0 {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (o *Order) getAllRunFuns() *[]string {
	return &RunFunList
}

func (o *Order) Gmv(c *gin.Context) {
	var result []order.Gmv
	sql := "SELECT\n\t\tproject_name,\n\t\tSUM(oi.goods_amount) AS goods_amount\n\tFROM\n\t\torder_info oi\n\tWHERE \n\t\toi.order_time BETWEEN '2022-01-01 00:00:00' AND '2022-12-20 00:00:00' \n\t\tAND pay_status IN (1,2)\n\t\tAND oi.email NOT LIKE '%@tetx.com' \n\t\tAND oi.email NOT LIKE '%@i9i8.com' \n\t\tAND oi.email NOT LIKE '%@qq.com' \n\t\tAND oi.email NOT LIKE '%@163.com' \n\t\tAND oi.email NOT LIKE '%@jjshouse.com' \n\t\tAND oi.email NOT LIKE '%@jenjenhouse.com'\nGROUP BY oi.project_name"
	o.DB.Raw(sql).Scan(&result)
	common.OkWithDetailed(result, "success", c)
	//fmt.Println(result)
	//atomic.AddInt64(remainCount, -1)
}

func (o *Order) TestGmv(remainCount *int64) {
	atomic.AddInt64(remainCount, -1)
	fmt.Println("TestGmv")
}
