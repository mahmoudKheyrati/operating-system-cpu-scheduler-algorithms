package schedulers

import (
	"log"
	"os-project/internal/core"
	"os-project/internal/requests"
	"os-project/internal/responses"
	"sync"
	"time"
)

func ScheduleMultilevelFeedbackQueue(request *requests.ScheduleRequests, timeQuantumList []int) (responses.ScheduleResponse, error) {
	log.Println("mlfq algorithm with timeQuantum = ", timeQuantumList)
	// we have 3 level. 5, 8, fcfs
	var ioDeviceCount = len(request.Jobs)

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan core.Proccess)
	var ioWorkQueue = make(chan core.Proccess, len(request.Jobs))
	var ioDoneQueue = make(chan core.Proccess, len(request.Jobs))
	var completedProcesses = make(chan core.Proccess)

	roundRobinChannel1 := make(chan core.Proccess, len(request.Jobs))
	roundRobinChannel2 := make(chan core.Proccess, len(request.Jobs))
	fcfsChannel := make(chan core.Proccess, len(request.Jobs))

	sendProccessToNextChannel := func(proccess core.Proccess) {
		if int(proccess.TimeQuantum.Seconds()) == timeQuantumList[0] {
			roundRobinChannel1 <- proccess // send proccess to next channel
		} else if int(proccess.TimeQuantum.Seconds()) == timeQuantumList[1] {
			roundRobinChannel2 <- proccess
		} else if int(proccess.TimeQuantum.Seconds()) == timeQuantumList[2] {
			fcfsChannel <- proccess
		}
	}

	contextSwitch := make(chan core.Proccess)
	go func() {
		for process := range contextSwitch {
			if process.Job.CpuTime1 == -1 && process.Job.IoTime != -1 { // we have io request
				log.Println("pid:", process.Job.ProcessId, "send io request")
				ioWorkQueue <- process
			} else { // just context switch
				log.Println("pid:", process.Job.ProcessId, "context switch detected. send proccess to roundRobin channel")

				process.TimeQuantum = getNextTimeQuantum(process.TimeQuantum, timeQuantumList)
				addNewScheduleTimeToProccess(&process)
				sendProccessToNextChannel(process)
			}
		}
	}()

	go func() { // ioDone
		for proccess := range ioDoneQueue {
			log.Println("pid:", proccess.Job.ProcessId, "ioDone!")
			sendProccessToNextChannel(proccess)
			//roundRobinChannel2 <- proccess // send proccess to its related channel
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
					TimeQuantum:   time.Duration(timeQuantumList[0]) * time.Second, // we put proccess in the first roundRobin queue
				}

				select {
				case <-time.After(time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())):
					log.Println("pid:", proccess.Job.ProcessId, "send process to roundRobin1 channel")
					addNewScheduleTimeToProccess(&proccess)
					sendProccessToNextChannel(proccess)
				}
			}(job)
		}

	}(&wg)

	go func() { // schedule multilevel feedback queue.
		for {
			select { // priority select
			case proccess := <-roundRobinChannel1:
				log.Println("pid:", proccess.Job.ProcessId, "roundRobinChannel1 send process to cpuWorkQueue")
				cpuWorkQueue <- proccess
				continue
			default:
				select {
				case proccess := <-roundRobinChannel2:
					log.Println("pid:", proccess.Job.ProcessId, "roundRobinChannel2 send process to cpuWorkQueue")
					cpuWorkQueue <- proccess
					continue
				default:
					select {
					case proccess := <-fcfsChannel:
						log.Println("pid:", proccess.Job.ProcessId, "fcfsChannel send process to cpuWorkQueue")
						cpuWorkQueue <- proccess
						continue
					default:
						continue
					}
				}
			}
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
func getNextTimeQuantum(currentTimeQuantum time.Duration, timeQuantumList []int) time.Duration {
	timeQuantumDuration := int(currentTimeQuantum.Seconds())
	for i := 0; i < len(timeQuantumList)-1; i++ {
		if timeQuantumList[i] == timeQuantumDuration {
			return time.Duration(timeQuantumList[i+1]) * time.Second
		}
	}
	return time.Duration(timeQuantumList[len(timeQuantumList)-1]) * time.Second // for last timeQuantum returns lastTimeQuantum
}

func addNewScheduleTimeToProccess(proccess *core.Proccess) {
	proccess.ScheduleTimes = append(proccess.ScheduleTimes, core.ScheduleTime{
		Submission: time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second),
	})
}
