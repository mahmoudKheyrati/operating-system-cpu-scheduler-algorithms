package schedulers

import (
	"log"
	"os-project/internal/core"
	"os-project/internal/requests"
	"os-project/internal/responses"
	"sort"
	"sync"
	"time"
)

func ScheduleFirstComeFirstServe(request requests.ScheduleRequests) (responses.ScheduleResponse, error) {
	var ioDeviceCount = len(request.Jobs)

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan core.Proccess)
	var ioWorkQueue = make(chan core.Proccess)
	var completedProcesses = make(chan core.Proccess)

	var cpuMetrics = make([]*core.CpuMetric, 0, 0)

	contextSwitch := make(chan core.Proccess)
	go func() {
		for process := range contextSwitch { // we now when process executes its burst time we should send it for io-Device.
			log.Println("pid:", process.Job.ProcessId, "send io request.")
			ioWorkQueue <- process
		}
	}()

	var wg sync.WaitGroup
	wg.Add(cpuCoresCount + ioDeviceCount + proccessSchedulerGoroutineCount + completionProccessGoroutineCount)

	// by this technique we could have multi-core and multi-io device at once
	for i := 0; i < cpuCoresCount; i++ {
		cpuMetrics = append(cpuMetrics, &core.CpuMetric{})
		go core.CpuExecute(&wg, cpuWorkQueue, completedProcesses, contextSwitch, cpuMetrics[i])
	}

	for i := 0; i < ioDeviceCount; i++ {
		go core.IoExecute(&wg, ioWorkQueue, cpuWorkQueue)
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
			go func(job requests.Job) {

				proccess := core.Proccess{
					Job:           &job,
					ScheduleTimes: make([]core.ScheduleTime, 0, 0),
				}
				proccess.ScheduleTimes = append(proccess.ScheduleTimes, core.ScheduleTime{
					Submission: time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second),
				})
				select {
				case <-time.After(time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())):
					log.Println("pid:", proccess.Job.ProcessId, "send process to cpuWorkQueue")
					cpuWorkQueue <- proccess
				}
				log.Println("pid:", proccess.Job.ProcessId, "schedule proccess.")
			}(job)
		}

	}(&wg)

	// get completed proccess metrics
	proccessDetails := make([]responses.ProcessResponse, 0)

	go func(waitGroup *sync.WaitGroup) {
		defer waitGroup.Done()
		for process := range completedProcesses {

			details := generateProcessDetails(process)
			proccessDetails = append(proccessDetails, details)

			if len(proccessDetails) == len(request.Jobs) {
				close(cpuWorkQueue)
				close(ioWorkQueue)
				break
			}
		}
	}(&wg)

	wg.Wait()

	response := generateResponse(len(request.Jobs), proccessDetails, cpuMetrics)
	log.Printf("response is: %+v", response)
	return response, nil
}
