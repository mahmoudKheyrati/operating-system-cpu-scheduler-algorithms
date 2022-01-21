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

	var startTime = time.Now()
	var utilizationTime = 0
	for proccess := range cpuWorkQueue {
		if proccess.Job.CpuTime1 != -1 {
			// execute cpu time 1
			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Execution = time.Now() // set execution time
			log.Println("cpu-time1 takes ", proccess.Job.CpuTime1, " seconds.")
			time.Sleep(time.Duration(proccess.Job.CpuTime1) * time.Second) // simulate execution
			utilizationTime += proccess.Job.CpuTime1
			proccess.Job.CpuTime1 = -1
			log.Println("cpu-time1 executes successfully. ")
			// context switch

			//go func() { // runs on another coroutine to ensure not waiting for request io
			log.Println("send io request. process: ", proccess)
			ioWorkQueue <- proccess
			//}()

		} else if proccess.Job.CpuTime1 == -1 && proccess.Job.IoTime != -1 {
			// todo: if time-quantum not finished we can send io-request

			// run io in the io queue
			go func() { // runs on another coroutine to ensure not waiting for request io
				log.Println("send io request. process: ", proccess)
				ioWorkQueue <- proccess
			}()
			// context switch
		} else if proccess.Job.CpuTime2 != -1 {
			// execute cpu time 2

			proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Execution = time.Now() // set execution time
			log.Println("cpu-time2 takes ", proccess.Job.CpuTime2, " seconds.")

			time.Sleep(time.Duration(proccess.Job.CpuTime2) * time.Second) // simulate execution
			utilizationTime += proccess.Job.CpuTime2
			proccess.Job.CpuTime2 = -1
			// context switch

			// use goroutine to ensure sending to channel is non-blocking
			go func() {
				// last execution: when proccess complete its execution we send it to done channel
				log.Println("send proccess to completed proccess channel. proccess: ", proccess)
				completedProcesses <- proccess
			}()

			log.Println("cpu-time1 executes successfully. ")

		}
	}
	var totalTime = time.Now().Sub(startTime)
	var cpuIdleTime = totalTime - time.Duration(utilizationTime)*time.Second

	// assign metrics
	metric = &CpuMetric{
		TotalTime:       totalTime,
		UtilizationTime: time.Duration(utilizationTime) * time.Second,
		IdleTime:        cpuIdleTime,
	}
	log.Printf("cpu metric: %+v", metric)

}

func IoExecute(wg *sync.WaitGroup, ioWorkQueue chan Proccess, cpuWorkQueue chan Proccess) {
	log.Println("start io-device")
	defer wg.Done()
	for proccess := range ioWorkQueue {
		log.Println("io-request with proccess", proccess)
		log.Println("io-request takes ", proccess.Job.IoTime, " seconds.")
		time.Sleep(time.Duration(proccess.Job.IoTime) * time.Second)

		proccess.Job.IoTime = -1
		// submit proccess to cpu to execute
		var scheduleTime = ScheduleTime{
			Submission: time.Now(),
		}
		proccess.ScheduleTimes = append(proccess.ScheduleTimes, scheduleTime)
		// add this job to ready queue
		cpuWorkQueue <- proccess
		log.Println("io-request done, proccess: ", proccess)
	}
}
