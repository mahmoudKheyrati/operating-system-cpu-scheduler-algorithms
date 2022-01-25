package schedulers

import (
	"fmt"
	"log"
	"os-project/internal/core"
	"os-project/internal/requests"
	"os-project/internal/responses"
	"os-project/internal/util"
	"sort"
	"sync"
	"time"
)

type localProccess struct {
	*core.Proccess
	isContextSwitch bool
}

func ScheduleShortestJobFirst(request *requests.ScheduleRequests) (responses.ScheduleResponse, error) {
	var ioDeviceCount = len(request.Jobs)

	var wg sync.WaitGroup
	wg.Add(cpuCoresCount + ioDeviceCount + proccessSchedulerGoroutineCount + completionProccessGoroutineCount)

	// initialize processes
	jobs := request.Jobs[:]
	processes := make([]*core.Proccess, 0)
	for i := 0; i < len(jobs); i++ {
		processes = append(processes, &core.Proccess{
			Job:           &jobs[i],
			ScheduleTimes: make([]core.ScheduleTime, 0, 0),
		})
	}
	printProcesses(processes)

	// run cpu cores and io devices
	var cpuWorkQueue = make(chan core.Proccess)
	var ioWorkQueue = make(chan core.Proccess)
	var ioDoneQueue = make(chan core.Proccess)
	var completedProcesses = make(chan core.Proccess)

	var contextSwitch = make(chan core.Proccess)

	var cpuMetrics = make([]*core.CpuMetric, 0, 0)

	// get completed proccess metrics
	proccessDetails := make([]responses.ProcessResponse, 0)

	go func(waitGroup *sync.WaitGroup) {
		defer waitGroup.Done()
		for process := range completedProcesses {
			responseTime := process.ScheduleTimes[0].Execution.Sub(process.ScheduleTimes[0].Submission)
			log.Println("pid:", process.Job.ProcessId, "proccess completed. schedule times: ", process.ScheduleTimes)
			turnAroundTime := process.ScheduleTimes[len(process.ScheduleTimes)-1].Complete.Sub(process.ScheduleTimes[0].Submission)

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

			if len(proccessDetails) == len(request.Jobs) {
				close(cpuWorkQueue)
				close(ioWorkQueue)
				close(ioDoneQueue)
				close(contextSwitch)
				break
			}
		}
	}(&wg)

	sortShortest := func() {
		sort.SliceStable(processes, func(i, j int) bool {

			if processes[i].Job.CpuTime1 != -1 && processes[j].Job.CpuTime1 != -1 {
				if processes[i].Job.ArrivalTime < processes[j].Job.ArrivalTime {
					return processes[i].Job.ArrivalTime < processes[j].Job.ArrivalTime
				}
			}

			if processes[i].Job.CpuTime1 != -1 && processes[j].Job.CpuTime1 != -1 {
				return processes[i].Job.CpuTime1 < processes[j].Job.CpuTime1
			} else if processes[i].Job.CpuTime1 != -1 && processes[j].Job.CpuTime1 == -1 {
				return processes[i].Job.CpuTime1 < processes[j].Job.CpuTime2
			} else if processes[i].Job.CpuTime1 == -1 && processes[j].Job.CpuTime1 != -1 {
				return processes[i].Job.CpuTime2 < processes[j].Job.CpuTime1
			}
			return processes[i].Job.CpuTime2 < processes[j].Job.CpuTime2
		})

	}

	var startTime time.Time
	scheduleProccess := func(proccess *core.Proccess) {
		log.Printf("proccess to schedule %+v with scheduleTime: %v", *proccess.Job, proccess.ScheduleTimes)
		if len(proccess.ScheduleTimes) == 0 {
			proccess.ScheduleTimes = append(proccess.ScheduleTimes, core.ScheduleTime{
				Submission: time.Now().Add(time.Duration(proccess.Job.ArrivalTime) * time.Second),
			})
			log.Println("pid:", proccess.Job.ProcessId, "RUUUUNN for first time.", (startTime.Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())))
			if startTime.Add(time.Duration(proccess.Job.ArrivalTime)*time.Second).Sub(time.Now()) < 0 {
				log.Println("pid:", proccess.Job.ProcessId, "schedule for first time immediately.")
				log.Println("pid:", proccess.Job.ProcessId, "sending process to cpuWorkQueue immediately")
				cpuWorkQueue <- *proccess
				return
			}
			select {
			case <-time.After(startTime.Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())):
				log.Println("pid:", proccess.Job.ProcessId, "schedule for first time.")
				log.Println("pid:", proccess.Job.ProcessId, "sending process to cpuWorkQueue")
				cpuWorkQueue <- *proccess
			}
		} else {
			log.Println("pid:", proccess.Job.ProcessId, "schedule for " , len(proccess.ScheduleTimes), "times")
			log.Println("pid:", proccess.Job.ProcessId, "schedule times " , proccess.ScheduleTimes)
			log.Printf("%+v", proccess)

			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Complete = time.Now()

			log.Println("pid:", proccess.Job.ProcessId, "sending process to cpuWorkQueue")
			cpuWorkQueue <- *proccess
		}
	}

	var reSchedule = func() {
		if len(processes) > 0 {
			sortShortest()
			log.Println("pid:", processes[0].Job.ProcessId, "******************************************** select as shortest process to schedule.")
			scheduleProccess(processes[0])
			log.Printf("-------------------- removing %+v %+v ", processes[0].Job, processes[0].ScheduleTimes)
			processes = processes[1:]
			printProcesses(processes)
		} else {
			log.Println("((((((((((((((((((((((((((((((((((( processes are empty")
		}
	}

	go func() {
		v, ok := <-contextSwitch
		if !ok {
			log.Println("OOOOOOOOOOOOOOOOOOOOOOOOO context closed")
			return
		}
		// issue io-request
		go func() {
			ioWorkQueue <- v
		}()

		// reschedule
		log.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ context switch issue reschedule")
		reSchedule()
	}()

	go func() {
		v, ok := <-ioDoneQueue
		if !ok {
			log.Println("OOOOOOOOOOOOOOOOOOOOOOOOO ioDoneQueue closed")
			return
		}

		log.Println("pid:", v.Job.ProcessId, "******************************** receive proccess in reschedule proccess")
		processes = append(processes, &v)
		log.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ ioDone issue reschedule")
		printProcesses(processes)
		reSchedule()
	}()

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

		log.Println("Schedule first process.")
		printProcesses(processes)
		startTime=time.Now()
		reSchedule()
		printProcesses(processes)

	}(&wg)

	wg.Wait()

	averageWaitingTime, averageResponseTime, averageTimeAroundTime := util.CalculateAverage(proccessDetails)

	utilization := 1 - (cpuMetrics[0].IdleTime.Seconds() / cpuMetrics[0].TotalTime.Seconds())
	var jobCount = len(request.Jobs)
	var response = responses.ScheduleResponse{
		TotalTime:             cpuMetrics[0].TotalTime.Seconds(),
		IdleTime:              cpuMetrics[0].IdleTime.Seconds(),
		CpuUtilization:        utilization,
		CpuThroughput:         float64(jobCount) / cpuMetrics[0].TotalTime.Seconds(),
		AverageWaitingTime:    averageWaitingTime,
		AverageResponseTime:   averageResponseTime,
		AverageTurnAroundTime: averageTimeAroundTime,
		Details:               proccessDetails,
	}

	log.Printf("response is: %+v", response)
	return response, nil
}

func printProcesses(processes []*core.Proccess) {
	var ids = make([]int, 0)
	for i := 0; i < len(processes); i++ {
		ids = append(ids, processes[i].Job.ProcessId)

	}
	fmt.Println(" $$$$$$$$$$$$$$$$$$ ids:", ids, "$$$$$$$$$$$$$$$$$$$4")
}
