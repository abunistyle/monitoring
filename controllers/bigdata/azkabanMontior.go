package bigdata

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"net/url"
	"time"
)

//数据库连接信息
const (
	USERNAME = "bigdata_ue2"
	PASSWORD = "bjQdmVLm732z"
	NETWORK  = "tcp"
	SERVER   = "43.198.52.10"
	PORT     = 6033
	DATABASE = "azkaban"
)

//azkaban任务
type ExecutionJobs struct {
	flowJob string `json:"flowJob" form:"flowJob"`
	status  int    `json:"status" form:"status"`
}

type AzkabanMotior struct {
}

func (a *AzkabanMotior) Sendnotice() {
	conn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("connection to mysql failed:", err)
		return
	} else {
		println("连接成功！")
	}

	//获取当日时间-北京时间
	currentTime := time.Now().Format("2006-01-02 15")
	fmt.Println(currentTime)
	//
	//sql :=fmt.Sprintf("select flow_id,status from azkaban.execution_jobs where '2022-09-01'='%s' limit 1",currentTime)
	sql := "select flow_id,status " +
		"from azkaban.execution_jobs " +
		"where DATE_FORMAT(date_add(from_unixtime(end_time/1000,'%Y-%m-%d %H'),interval 7 hour),'%Y-%m-%d %H')<=? " +
		"and DATE_FORMAT(date_add(from_unixtime(end_time/1000,'%Y-%m-%d %H'),interval 8 hour),'%Y-%m-%d %H')>=? " +
		"and status in (50,60)"
	rows, err := db.Query(sql, currentTime, currentTime)

	if err != nil {
		panic(err)
		return
	}

	//接受sql查询结果对象
	executionJobs := new(ExecutionJobs)

	for rows.Next() {
		row := rows.Scan(&executionJobs.flowJob, &executionJobs.status)
		if row != nil {
			fmt.Println("数据为空")
		}

		if executionJobs.status == 50 {
			message := "azkaban调度" + "\nazkaban任务：" + executionJobs.flowJob + "\n任务状态：failed"
			fmt.Println(message)
			resp, err := http.Get(fmt.Sprintf("http://voice.arch800.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s", url.QueryEscape(message)))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(resp)
		}

		if executionJobs.status == 60 {
			message := "\nazkaban任务：" + executionJobs.flowJob + "\n任务状态：killed"
			fmt.Println(message)
			resp, err := http.Get(fmt.Sprintf("http://voice.arch800.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s", url.QueryEscape(message)))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(resp)
		}

	}

	db.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超时的连接就close
	db.SetMaxOpenConns(10)                   //设置最大连接数

}
