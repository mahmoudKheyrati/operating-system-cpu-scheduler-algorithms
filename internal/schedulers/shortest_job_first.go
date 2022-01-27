package schedulers

import (
	"fmt"
	"log"
	"os-project/internal/core"
	"os-project/internal/requests"
	"os-project/internal/responses"
	"sort"
	"sync"
	"time"
)

func ScheduleShortestJobFirst(request *requests.ScheduleRequests) (responses.ScheduleResponse, error) {
	log.Println("running sjf algorithm ...")
	// we have 3 level. 5, 8, fcfs
	var ioDeviceCount = len(request.Jobs)

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan core.Proccess)
	var ioWorkQueue = make(chan core.Proccess, len(request.Jobs))
	var ioDoneQueue = make(chan core.Proccess, len(request.Jobs))
	var completedProcesses = make(chan core.Proccess)

	readyQueue := make([]core.Proccess, 0)
	var readyQueueMutex sync.Mutex

	scheduleNewProccess := func() {
		time.Sleep(20 * time.Millisecond)
		readyQueueMutex.Lock()
		defer readyQueueMutex.Unlock()
		// sort ready queue
		readyQueue = sortShortestJob(readyQueue)
		// pick shortest
		if len(readyQueue) > 0 {
			shortestJob := readyQueue[0]
			readyQueue = readyQueue[1:]

			select { // try to send the shortest job to cpu
			case cpuWorkQueue <- shortestJob:
			default:
				readyQueue = append(readyQueue, shortestJob)
			}
		}
	}

	addProccessToReadyQueue := func(proccess core.Proccess) {
		readyQueueMutex.Lock()
		readyQueue = append(readyQueue, proccess)
		readyQueueMutex.Unlock()

		scheduleNewProccess()
	}

	contextSwitch := make(chan core.Proccess)
	go func() {
		for process := range contextSwitch {
			if process.Job.CpuTime1 == -1 && process.Job.IoTime != -1 { // we have io request
				go scheduleNewProccess()
				log.Println("pid:", process.Job.ProcessId, "send io request")
				ioWorkQueue <- process
			}
		}
	}()

	go func() { // ioDone
		for proccess := range ioDoneQueue {
			log.Println("pid:", proccess.Job.ProcessId, "ioDone!")
			addProccessToReadyQueue(proccess)
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
		go core.IoExecute(&wg, ioWorkQueue, ioDoneQueue)
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
				}

				select {
				case <-time.After(time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())):
					addNewScheduleTimeToProccess(&proccess, time.Now())
					log.Println("pid:", proccess.Job.ProcessId, "send process to readyQueue")
					addProccessToReadyQueue(proccess)
				}
			}(job)
		}
	}(&wg)

	go func() { // schedule multilevel feedback queue.
		//for {
		//	cpuWorkQueue <- proccess
		//}
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

func sortShortestJob(processes []core.Proccess) []core.Proccess {
	sort.SliceStable(processes, func(i, j int) bool {

		//if processes[i].Job.CpuTime1 != -1 && processes[j].Job.CpuTime1 != -1 {
		//	if processes[i].Job.ArrivalTime < processes[j].Job.ArrivalTime {
		//		return processes[i].Job.ArrivalTime < processes[j].Job.ArrivalTime
		//	}
		//}

		if processes[i].Job.CpuTime1 != -1 && processes[j].Job.CpuTime1 != -1 {
			return processes[i].Job.CpuTime1 < processes[j].Job.CpuTime1
		} else if processes[i].Job.CpuTime1 != -1 && processes[j].Job.CpuTime1 == -1 {
			return processes[i].Job.CpuTime1 < processes[j].Job.CpuTime2
		} else if processes[i].Job.CpuTime1 == -1 && processes[j].Job.CpuTime1 != -1 {
			return processes[i].Job.CpuTime2 < processes[j].Job.CpuTime1
		}
		return processes[i].Job.CpuTime2 < processes[j].Job.CpuTime2
	})
	fmt.Printf("sorted proccesses: ")
	for i := 0; i < len(processes); i++ {
		fmt.Printf("job: %+v, ", processes[i].Job)
	}
	fmt.Println()
	return processes
}
