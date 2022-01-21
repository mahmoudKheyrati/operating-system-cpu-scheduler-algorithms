package schedulers

import (
	"os-project/internal/requests"
	"os-project/internal/responses"
	"sort"
	"sync"
	"time"
)

type proccess struct {
	responses.ProcessResponse
}

type cpuCore struct {
}

type cpuMetric struct {
	TotalTime       time.Duration
	UtilizationTime time.Duration
	IdleTime        time.Duration
}

func cpuExecute(wg *sync.WaitGroup, cpuWorkQueue chan requests.Job, ioWorkQueue chan requests.Job, metric *cpuMetric) {
	defer wg.Done()
	var startTime = time.Now()
	var utilizationTime = 0
	for job := range cpuWorkQueue {
		if job.CpuTime1 != -1 {
			// execute cpu time 1
			time.Sleep(time.Duration(job.CpuTime1) * time.Second) // simulate execution
			job.CpuTime1 = -1
			utilizationTime += job.CpuTime1
			utilizationTime += job.CpuTime1
			// context switch

		} else if job.CpuTime1 == -1 && job.IoTime != -1 {
			// run io in the io queue
			go func() { // runs on another coroutine to ensure not waiting for request io
				ioWorkQueue <- job
			}()
			// context switch
		} else if job.CpuTime2 != -1 {
			// execute cpu time 2

			time.Sleep(time.Duration(job.CpuTime2) * time.Second) // simulate execution
			job.CpuTime2 = -1
			utilizationTime += job.CpuTime2
			// context switch
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

func ioExecute(wg *sync.WaitGroup, ioWorkQueue chan requests.Job, cpuWorkQueue chan requests.Job) {
	defer wg.Done()
	for job := range ioWorkQueue {
		time.Sleep(time.Duration(job.IoTime) * time.Second)
		job.IoTime = -1
		// add this job to ready queue
		cpuWorkQueue <- job
	}
}

func ScheduleFirstComeFirstServe(request requests.ScheduleRequests) (*responses.ScheduleResponse, error) {
	var cpuCoresCount = 1
	var ioDeviceCount = 1

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan requests.Job)
	var ioWorkQueue = make(chan requests.Job)

	var cpuMetrics = make([]*cpuMetric, 0, 0)

	var wg sync.WaitGroup
	wg.Add(cpuCoresCount + ioDeviceCount + 1)

	// by this technique we could have multi-core and multi-io device at once
	for i := 0; i < cpuCoresCount; i++ {
		cpuMetrics = append(cpuMetrics, &cpuMetric{})
		go cpuExecute(&wg, cpuWorkQueue, ioWorkQueue, cpuMetrics[i])
	}

	for i := 0; i < ioDeviceCount; i++ {
		go ioExecute(&wg, cpuWorkQueue, ioWorkQueue)
	}

	// schedule jobs
	go func(waitGroup *sync.WaitGroup) {
		defer waitGroup.Done()

		// sort based on arrival time
		jobs := request.Jobs[:]
		sort.SliceStable(jobs, func(i, j int) bool {
			return jobs[i].ArrivalTime < jobs[j].ArrivalTime
		})

		// schedule jobs
		for _, job := range jobs {
			cpuWorkQueue <- job
		}
		close(cpuWorkQueue)
		close(ioWorkQueue)

	}(&wg)

	wg.Wait()

	utilization := 1 - (cpuMetrics[0].IdleTime.Seconds() / cpuMetrics[0].TotalTime.Seconds())
	var jobCount = len(request.Jobs)
	var response = &responses.ScheduleResponse{
		TotalTime:             cpuMetrics[0].TotalTime.Seconds(),
		IdleTime:              cpuMetrics[0].IdleTime.Seconds(),
		CpuUtilization:        utilization,
		CpuThroughput:         float64(jobCount) / cpuMetrics[0].TotalTime.Seconds(),
		AverageWaitingTime:    0,
		AverageResponseTime:   0,
		AverageTurnAroundTime: 0,
		Details: []responses.ProcessResponse{
			{
				ProcessId:      0,
				ResponseTime:   0,
				TurnAroundTime: 0,
				WaitingTime:    0,
			},
		},
	}

	return response, nil
}
