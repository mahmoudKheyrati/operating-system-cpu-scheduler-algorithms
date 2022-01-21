package schedulers

import (
	"os-project/internal/requests"
	"os-project/internal/responses"
	"sort"
	"sync"
	"time"
)

type cpuCore struct {
}

type scheduleTime struct {
	submission time.Time
	execution  time.Time
}
type Proccess struct {
	job           *requests.Job
	ScheduleTimes []scheduleTime
}

type cpuMetric struct {
	TotalTime       time.Duration
	UtilizationTime time.Duration
	IdleTime        time.Duration
}

func cpuExecute(wg *sync.WaitGroup, cpuWorkQueue chan Proccess, ioWorkQueue chan Proccess, completedProcesses chan Proccess, metric *cpuMetric) {
	defer wg.Done()
	defer close(completedProcesses)

	var startTime = time.Now()
	var utilizationTime = 0
	for proccess := range cpuWorkQueue {
		if proccess.job.CpuTime1 != -1 {
			// execute cpu time 1
			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].execution = time.Now() // set execution time
			time.Sleep(time.Duration(proccess.job.CpuTime1) * time.Second)               // simulate execution
			proccess.job.CpuTime1 = -1
			utilizationTime += proccess.job.CpuTime1
			utilizationTime += proccess.job.CpuTime1
			// context switch

		} else if proccess.job.CpuTime1 == -1 && proccess.job.IoTime != -1 {
			// todo: if time-quantum not finished we can send io-request

			// run io in the io queue
			go func() { // runs on another coroutine to ensure not waiting for request io
				ioWorkQueue <- proccess
			}()
			// context switch
		} else if proccess.job.CpuTime2 != -1 {
			// execute cpu time 2

			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].execution = time.Now() // set execution time
			time.Sleep(time.Duration(proccess.job.CpuTime2) * time.Second)               // simulate execution
			proccess.job.CpuTime2 = -1
			utilizationTime += proccess.job.CpuTime2
			// context switch

			// use goroutine to ensure sending to channel is non-blocking
			go func() {
				// last execution: when proccess complete its execution we send it to done channel
				completedProcesses <- proccess
			}()

		}
	}
	var totalTime = time.Now().Sub(startTime)
	var cpuIdleTime = totalTime - time.Duration(utilizationTime)*time.Second

	// assign metrics
	metric = &cpuMetric{
		TotalTime:       totalTime,
		UtilizationTime: time.Duration(utilizationTime) * time.Second,
		IdleTime:        cpuIdleTime,
	}

}

func ioExecute(wg *sync.WaitGroup, ioWorkQueue chan Proccess, cpuWorkQueue chan Proccess) {
	defer wg.Done()
	for proccess := range ioWorkQueue {
		time.Sleep(time.Duration(proccess.job.IoTime) * time.Second)
		proccess.job.IoTime = -1
		// submit proccess to cpu to execute
		var scheduleTime = scheduleTime{
			submission: time.Now(),
		}
		proccess.ScheduleTimes = append(proccess.ScheduleTimes, scheduleTime)
		// add this job to ready queue
		cpuWorkQueue <- proccess
	}
}

func ScheduleFirstComeFirstServe(request requests.ScheduleRequests) (*responses.ScheduleResponse, error) {
	const cpuCoresCount = 1
	var ioDeviceCount = len(request.Jobs)
	const proccessSchedulerGoroutineCount = 1
	const completionProccessGoroutineCount = 1

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan Proccess)
	var ioWorkQueue = make(chan Proccess)
	var completedProcesses = make(chan Proccess)

	var cpuMetrics = make([]*cpuMetric, 0, 0)

	var wg sync.WaitGroup
	wg.Add(cpuCoresCount + ioDeviceCount + proccessSchedulerGoroutineCount + completionProccessGoroutineCount)

	// by this technique we could have multi-core and multi-io device at once
	for i := 0; i < cpuCoresCount; i++ {
		cpuMetrics = append(cpuMetrics, &cpuMetric{})
		go cpuExecute(&wg, cpuWorkQueue, ioWorkQueue, completedProcesses, cpuMetrics[i])
	}

	for i := 0; i < ioDeviceCount; i++ {
		go ioExecute(&wg, cpuWorkQueue, ioWorkQueue)
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
			proccess := Proccess{
				job:           &job,
				ScheduleTimes: make([]scheduleTime, 0, 0),
			}
			proccess.ScheduleTimes = append(proccess.ScheduleTimes, scheduleTime{
				submission: time.Now(),
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
			responseTime := process.ScheduleTimes[0].execution.Sub(process.ScheduleTimes[0].submission)
			turnAroundTime := process.ScheduleTimes[len(process.ScheduleTimes)-1].execution.Sub(process.ScheduleTimes[0].submission)

			var waitingTime float64 = 0
			for _, s := range process.ScheduleTimes {
				waitingTime += s.execution.Sub(s.submission).Seconds()
			}

			proccessDetails = append(proccessDetails, responses.ProcessResponse{
				ProcessId:      process.job.ProcessId,
				ResponseTime:   responseTime.Seconds(),
				TurnAroundTime: turnAroundTime.Seconds(),
				WaitingTime:    waitingTime,
			})
		}
	}(&wg)

	wg.Wait()
	averageWaitingTime, averageResponseTime, averageTimeAroundTime := calculateAverage(proccessDetails)

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

func calculateAverage(proccessDetails []responses.ProcessResponse) (averageWaitingTime, averageResponseTime, averageTimeAroundTime float64) {
	var waitingTimeSum float64
	var responseTimeSum float64
	var turnAroundTimeSum float64

	for _, proccess := range proccessDetails {
		waitingTimeSum += proccess.WaitingTime
		responseTimeSum += proccess.ResponseTime
		turnAroundTimeSum += proccess.TurnAroundTime
	}

	proccessCount := float64(len(proccessDetails))

	averageWaitingTime = waitingTimeSum / proccessCount
	averageResponseTime = responseTimeSum / proccessCount
	averageTimeAroundTime = turnAroundTimeSum / proccessCount
	return
}
