package requests

type Job struct {
	ProcessId   int `json:"process_id"`
	ArrivalTime int `json:"arrival_time"`
	CpuTime1    int `json:"cpu_time_1"`
	IoTime      int `json:"io_time"`
	CpuTime2    int `json:"cpu_time_2"`
}
type ScheduleRequests struct {
	Jobs []Job `json:"jobs"`
}
