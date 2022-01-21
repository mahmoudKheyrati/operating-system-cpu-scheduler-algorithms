package schedulers

import (
	"os-project/internal/core"
	"os-project/internal/requests"
	"os-project/internal/responses"
	"os-project/internal/util"
	"sort"
	"sync"
	"time"
)

const (
	cpuCoresCount                    = 1
	proccessSchedulerGoroutineCount  = 1
	completionProccessGoroutineCount = 1
)

type cpuCore struct {
}

func ScheduleFirstComeFirstServe(request requests.ScheduleRequests) (*responses.ScheduleResponse, error) {
	var ioDeviceCount = len(request.Jobs)

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan core.Proccess)
	var ioWorkQueue = make(chan core.Proccess)
	var completedProcesses = make(chan core.Proccess)

	var cpuMetrics = make([]*core.CpuMetric, 0, 0)

	var wg sync.WaitGroup
	wg.Add(cpuCoresCount + ioDeviceCount + proccessSchedulerGoroutineCount + completionProccessGoroutineCount)

	// by this technique we could have multi-core and multi-io device at once
	for i := 0; i < cpuCoresCount; i++ {
		cpuMetrics = append(cpuMetrics, &core.CpuMetric{})
		go core.CpuExecute(&wg, cpuWorkQueue, ioWorkQueue, completedProcesses, cpuMetrics[i])
	}

	for i := 0; i < ioDeviceCount; i++ {
		go core.IoExecute(&wg, cpuWorkQueue, ioWorkQueue)
	}

	// schedule jobs
	go func(waitGroup *sync.WaitGroup) {
		defer waitGroup.Done()

		// sort jobs by arrival time
		jobs := request.Jobs[:]
		sort.SliceStable(jobs, func(i, j int) bool {
			return jobs[i].ArrivalTime < jobs[j].ArrivalTime
		})

		// schedule jobs
		for _, job := range jobs {
			proccess := core.Proccess{
				Job:           &job,
				ScheduleTimes: make([]core.ScheduleTime, 0, 0),
			}
			proccess.ScheduleTimes = append(proccess.ScheduleTimes, core.ScheduleTime{
				Submission: time.Now(),
			})
			cpuWorkQueue <- proccess
		}
		close(cpuWorkQueue)
		close(ioWorkQueue)

	}(&wg)

	// get completed proccess metrics
	proccessDetails := make([]responses.ProcessResponse, 0)

	go func(waitGroup *sync.WaitGroup) {
		defer waitGroup.Done()
		for process := range completedProcesses {
			responseTime := process.ScheduleTimes[0].Execution.Sub(process.ScheduleTimes[0].Submission)
			turnAroundTime := process.ScheduleTimes[len(process.ScheduleTimes)-1].Execution.Sub(process.ScheduleTimes[0].Submission)

			var waitingTime float64 = 0
			for _, s := range process.ScheduleTimes {
				waitingTime += s.Execution.Sub(s.Submission).Seconds()
			}

			proccessDetails = append(proccessDetails, responses.ProcessResponse{
				ProcessId:      process.Job.ProcessId,
				ResponseTime:   responseTime.Seconds(),
				TurnAroundTime: turnAroundTime.Seconds(),
				WaitingTime:    waitingTime,
			})
		}
	}(&wg)

	wg.Wait()
	averageWaitingTime, averageResponseTime, averageTimeAroundTime := util.CalculateAverage(proccessDetails)

	utilization := 1 - (cpuMetrics[0].IdleTime.Seconds() / cpuMetrics[0].TotalTime.Seconds())
	var jobCount = len(request.Jobs)
	var response = &responses.ScheduleResponse{
		TotalTime:             cpuMetrics[0].TotalTime.Seconds(),
		IdleTime:              cpuMetrics[0].IdleTime.Seconds(),
		CpuUtilization:        utilization,
		CpuThroughput:         float64(jobCount) / cpuMetrics[0].TotalTime.Seconds(),
		AverageWaitingTime:    averageWaitingTime,
		AverageResponseTime:   averageResponseTime,
		AverageTurnAroundTime: averageTimeAroundTime,
		Details:               proccessDetails,
	}

	return response, nil
}
