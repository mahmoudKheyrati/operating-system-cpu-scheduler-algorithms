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
	TimeQuantum   time.Duration
}

type CpuMetric struct {
	TotalTime       time.Duration
	UtilizationTime time.Duration
	IdleTime        time.Duration
}

func CpuExecute(wg *sync.WaitGroup, cpuWorkQueue chan Proccess, completedProcesses chan Proccess, contextSwitch chan Proccess, metric *CpuMetric) {
	log.Println("start cpu")
	defer wg.Done()
	defer close(completedProcesses)

	var startTime = time.Now()
	var utilizationTime = time.Duration(0)
	for proccess := range cpuWorkQueue {
		if proccess.Job.CpuTime1 != -1 { // execute cpu time 1
			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Execution = time.Now() // set execution time

			var cpuTime1Duration = time.Duration(proccess.Job.CpuTime1) * time.Second
			var timeToSleep = cpuTime1Duration
			var nextCpuTime1 = -1
			if proccess.TimeQuantum > 0 {

				if cpuTime1Duration >= proccess.TimeQuantum { // proccess bigger to execute in one cpu timeQuantum
					timeToSleep = proccess.TimeQuantum
					nextCpuTime1 = int((cpuTime1Duration - proccess.TimeQuantum).Seconds())
					if nextCpuTime1 <= 0 {
						nextCpuTime1 = -1
					}
					//log.Println("cputTime1Duration: ", cpuTime1Duration, "proccess.TimeQuantum: ", proccess.TimeQuantum, "= ", cpuTime1Duration-proccess.TimeQuantum)
					//log.Println("pid:", proccess.Job.ProcessId, "nextCputTime1", nextCpuTime1)
				}

			}

			log.Println("pid:", proccess.Job.ProcessId, "###################################### start executing in cpuCore1")
			time.Sleep(timeToSleep) // simulate execution
			log.Println("pid:", proccess.Job.ProcessId, "cpu-time1 takes ", timeToSleep)
			utilizationTime += timeToSleep
			proccess.Job.CpuTime1 = nextCpuTime1
			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Complete = time.Now()
			log.Println("pid:", proccess.Job.ProcessId, "cpu-time1 executes successfully. ")

			// context switch
			log.Printf("proccess to context switch: %+v", proccess.Job)

			contextSwitch <- proccess
			log.Println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% after cpu context switch")

			//} else if proccess.Job.CpuTime1 == -1 && proccess.Job.IoTime != -1 {
			//	// todo: if time-quantum not finished we can send io-request
			//
			//	// run io in the io queue
			//	//go func() { // runs on another coroutine to ensure not waiting for request io
			//	//	log.Println("pid:", proccess.Job.ProcessId,"send io request.")
			//	//	ioWorkQueue <- proccess
			//	//}()
			//	// context switch
			//}
		} else if proccess.Job.CpuTime2 != -1 {
			// execute cpu time 2

			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Execution = time.Now() // set execution time

			var cpuTime2Duration = time.Duration(proccess.Job.CpuTime2) * time.Second
			var timeToSleep = cpuTime2Duration
			var nextCpuTime2 = -1
			if proccess.TimeQuantum > 0 {
				if cpuTime2Duration >= proccess.TimeQuantum { // proccess bigger to execute in one cpu timeQuantum
					timeToSleep = proccess.TimeQuantum
					nextCpuTime2 = int((cpuTime2Duration - proccess.TimeQuantum).Seconds())
					if nextCpuTime2 <= 0 {
						nextCpuTime2 = -1
					}
				}
			}

			log.Println("pid:", proccess.Job.ProcessId, "start executing in cpuCore2")
			time.Sleep(timeToSleep) // simulate execution
			log.Println("pid:", proccess.Job.ProcessId, "cpu-time2 takes ", timeToSleep)
			utilizationTime += timeToSleep
			proccess.Job.CpuTime2 = nextCpuTime2
			// context switch

			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Complete = time.Now()
			log.Println("pid:", proccess.Job.ProcessId, "cpu-time2 executes successfully. ")

			if nextCpuTime2 == -1 { // last execution: when proccess complete its execution we send it to done channel
				log.Println("pid:", proccess.Job.ProcessId, "send proccess to completed proccess channel.")
				completedProcesses <- proccess
			} else { // process doesn't complete yet, so we should perform context switch
				log.Printf("proccess to context switch: %+v", proccess)
				contextSwitch <- proccess
			}

		}
	}
	var totalTime = time.Now().Sub(startTime)
	var cpuIdleTime = totalTime - utilizationTime

	// assign metrics
	*metric = CpuMetric{
		TotalTime:       totalTime,
		UtilizationTime: utilizationTime,
		IdleTime:        cpuIdleTime,
	}
	log.Printf("cpu metric: %+v", metric)

}
