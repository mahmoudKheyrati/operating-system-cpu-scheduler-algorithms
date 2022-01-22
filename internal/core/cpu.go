package core

import (
	"log"
	"os-project/internal/requests"
	"sync"
	"time"
)

type ScheduleTime struct {
	Submission time.Time
	Execution  time.Time
	Complete   time.Time
}
type Proccess struct {
	Job           *requests.Job
	ScheduleTimes []ScheduleTime
}

type CpuMetric struct {
	TotalTime       time.Duration
	UtilizationTime time.Duration
	IdleTime        time.Duration
}

func CpuExecute(wg *sync.WaitGroup, cpuWorkQueue chan Proccess, ioWorkQueue chan Proccess, completedProcesses chan Proccess, metric *CpuMetric) {
	log.Println("start cpu")
	defer wg.Done()
	defer close(completedProcesses)

	var utilizationTime = 0
	var startTime = time.Now()
	for proccess := range cpuWorkQueue {
		// arrival time in the future
		//nextArrivalDuration := startTime.Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())
		//log.Println("pid:", proccess.Job.ProcessId, "next arrival duration: ", nextArrivalDuration)
		//if nextArrivalDuration > 0 {
		//	log.Println("pid: ", proccess.Job.ProcessId, "delay")
		//	go func(p Proccess) {
		//		select {
		//		case <-time.After(startTime.Add(time.Duration(proccess.Job.ArrivalTime) * time.Second).Sub(time.Now())):
		//			// re-schedule
		//			cpuWorkQueue <- proccess
		//		}
		//	}(proccess)
		//	continue
		//}

		if proccess.Job.CpuTime1 != -1 {
			// execute cpu time 1
			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Execution = time.Now() // set execution time
			log.Println("pid:", proccess.Job.ProcessId, "cpu-time1 takes ", proccess.Job.CpuTime1, " seconds.")
			time.Sleep(time.Duration(proccess.Job.CpuTime1) * time.Second) // simulate execution
			utilizationTime += proccess.Job.CpuTime1
			proccess.Job.CpuTime1 = -1
			log.Println("pid:", proccess.Job.ProcessId, "cpu-time1 executes successfully. ")

			log.Println("pid:", proccess.Job.ProcessId, "send io request.")
			//go func() { // runs on another coroutine to ensure not waiting for request io
			// context switch
			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Complete = time.Now()
			ioWorkQueue <- proccess
			//}()

		} else if proccess.Job.CpuTime1 == -1 && proccess.Job.IoTime != -1 {
			// todo: if time-quantum not finished we can send io-request

			// run io in the io queue
			//go func() { // runs on another coroutine to ensure not waiting for request io
			//	log.Println("pid:", proccess.Job.ProcessId,"send io request.")
			//	ioWorkQueue <- proccess
			//}()
			// context switch
		} else if proccess.Job.CpuTime2 != -1 {
			// execute cpu time 2

			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Execution = time.Now() // set execution time

			log.Println("pid:", proccess.Job.ProcessId, "cpu-time2 takes ", proccess.Job.CpuTime2, " seconds.")

			time.Sleep(time.Duration(proccess.Job.CpuTime2) * time.Second) // simulate execution
			utilizationTime += proccess.Job.CpuTime2
			proccess.Job.CpuTime2 = -1
			// context switch

			log.Println("pid:", proccess.Job.ProcessId, "cpu-time2 executes successfully. ")
			// use goroutine to ensure sending to channel is non-blocking
			go func() {
				// last execution: when proccess complete its execution we send it to done channel
				log.Println("pid:", proccess.Job.ProcessId, "send proccess to completed proccess channel.")
				proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Complete = time.Now()
				completedProcesses <- proccess
			}()

		}
	}
	var totalTime = time.Now().Sub(startTime)
	var cpuIdleTime = totalTime - time.Duration(utilizationTime)*time.Second

	// assign metrics
	*metric = CpuMetric{
		TotalTime:       totalTime,
		UtilizationTime: time.Duration(utilizationTime) * time.Second,
		IdleTime:        cpuIdleTime,
	}
	log.Printf("cpu metric: %+v", metric)

}
