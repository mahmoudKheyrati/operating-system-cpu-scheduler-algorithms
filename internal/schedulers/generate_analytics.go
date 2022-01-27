package schedulers

import (
	"log"
	"os-project/internal/core"
	"os-project/internal/responses"
	"os-project/internal/util"
)

func generateResponse(processCount int, proccessDetails []responses.ProcessResponse, cpuMetrics []*core.CpuMetric) responses.ScheduleResponse {
	averageWaitingTime, averageResponseTime, averageTimeAroundTime := util.CalculateAverage(proccessDetails)

	utilization := 1 - (cpuMetrics[0].IdleTime.Seconds() / cpuMetrics[0].TotalTime.Seconds())
	var response = responses.ScheduleResponse{
		TotalTime:             cpuMetrics[0].TotalTime.Seconds(),
		IdleTime:              cpuMetrics[0].IdleTime.Seconds(),
		CpuUtilization:        utilization,
		CpuThroughput:         float64(processCount) / cpuMetrics[0].TotalTime.Seconds(),
		AverageWaitingTime:    averageWaitingTime,
		AverageResponseTime:   averageResponseTime,
		AverageTurnAroundTime: averageTimeAroundTime,
		Details:               proccessDetails,
	}
	return response
}

func generateProcessDetails(process core.Proccess) responses.ProcessResponse {
	responseTime := process.ScheduleTimes[0].Execution.Sub(process.ScheduleTimes[0].Submission)
	log.Println("pid:", process.Job.ProcessId, "proccess completed. schedule times: ", process.ScheduleTimes)
	turnAroundTime := process.ScheduleTimes[len(process.ScheduleTimes)-1].Complete.Sub(process.ScheduleTimes[0].Submission)

	var waitingTime float64 = 0
	for _, s := range process.ScheduleTimes {
		waitingTime += s.Execution.Sub(s.Submission).Seconds()
	}

	return responses.ProcessResponse{
		ProcessId:      process.Job.ProcessId,
		ResponseTime:   responseTime.Seconds(),
		TurnAroundTime: turnAroundTime.Seconds(),
		WaitingTime:    waitingTime,
	}
}
