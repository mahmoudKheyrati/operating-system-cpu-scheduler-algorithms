package responses

type ProcessResponse struct {
	ProcessId      int     `json:"process_id"`
	ResponseTime   float64 `json:"response_time"`
	TurnAroundTime float64 `json:"turn_around_time"`
	WaitingTime    float64 `json:"waiting_time"`
}
type ScheduleResponse struct {
	TotalTime             float64           `json:"total_time"`
	IdleTime              float64           `json:"idle_time"`
	AverageWaitingTime    float64           `json:"average_waiting_time"`
	AverageResponseTime   float64           `json:"average_response_time"`
	AverageTurnAroundTime float64           `json:"average_turn_around_time"`
	CpuUtilization        float64           `json:"cpu_utilization"`
	CpuThroughput         float64           `json:"cpu_throughput"`
	Details               []ProcessResponse `json:"details"`
}
