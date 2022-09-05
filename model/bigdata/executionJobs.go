package bigdata

type ExecutionJobs struct {
	flowJob string `json:"flowJob" 	form:"flowJob"`
	status  int    `json:"status" 	form:"status"`
}
