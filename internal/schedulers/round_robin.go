package schedulers

import (
	"log"
	"os-project/internal/core"
	"os-project/internal/requests"
	"os-project/internal/responses"
	"sync"
	"time"
)

//type ProcessQueue struct {
//	queue []core.Proccess
//	mu sync.Mutex
//}
//
//func NewProcessQueue() *ProcessQueue {
//	return &ProcessQueue{queue: make([]core.Proccess, 0)}
//}
//func (p *ProcessQueue) AddToEnd(proccess core.Proccess) {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	p.queue = append(p.queue, proccess)
//}
//func (p *ProcessQueue) RemoveFromTop() (core.Proccess, bool) {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	if len(p.queue) >  0  {
//		item := p.queue[0]
//		p.queue = p.queue[1:]
//		return item, true
//	}
//	return core.Proccess{}, false
//}

func ScheduleRoundRobin(request *requests.ScheduleRequests, timeQuantum int) (responses.ScheduleResponse, error) {
	log.Println("running roundRobin algorithm with timeQuantum = ", timeQuantum)
	var ioDeviceCount = len(request.Jobs)

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan core.Proccess)
	var ioWorkQueue = make(chan core.Proccess)
	var completedProcesses = make(chan core.Proccess)

	//roundRoubinQueue := NewProcessQueue()
	roundRobinChannel := make(chan core.Proccess, len(request.Jobs))

	contextSwitch := make(chan core.Proccess)
	go func() {
		for process := range contextSwitch {
			if process.Job.CpuTime1 == -1 && process.Job.IoTime != -1 { // we have io request
				log.Println("pid:", process.Job.ProcessId, "send io request")
				ioWorkQueue <- process
			} else { // just context switch
				log.Println("pid:", process.Job.ProcessId, "context switch detected. send proccess to roundRobin channel")
				roundRobinChannel <- process
			}
		}
	}()
	var cpuMetrics = make([]*core.CpuMetric, 0, 0)

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

		// schedule jobs
		for _, job := range jobs {
			go func(job requests.Job) {
				proccess := core.Proccess{
					Job:           &job,
					ScheduleTimes: make([]core.ScheduleTime, 0, 0),
					TimeQuantum:   time.Duration(timeQuantum) * time.Second,
				}
				proccess.ScheduleTimes = append(proccess.ScheduleTimes, core.ScheduleTime{
					Submission: time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second),
				})
				select {
				case <-time.After(time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())):
					log.Println("pid:", proccess.Job.ProcessId, "send process to roundRobin channel")
					roundRobinChannel <- proccess
				}
			}(job)
		}

	}(&wg)

	go func() {
		for proccess := range roundRobinChannel {
			log.Println("pid:", proccess.Job.ProcessId, "send process to cpuWorkQueue")
			cpuWorkQueue <- proccess
		}

	}()

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
