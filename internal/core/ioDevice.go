package core

import (
	"log"
	"sync"
	"time"
)

func IoExecute(wg *sync.WaitGroup, ioWorkQueue chan Proccess, cpuWorkQueue chan Proccess) {
	log.Println("start io-device")
	defer wg.Done()
	for proccess := range ioWorkQueue {
		log.Println("pid:", proccess.Job.ProcessId, "io-request performs")
		log.Println("pid:", proccess.Job.ProcessId, "io-request takes ", proccess.Job.IoTime, " seconds.")
		time.Sleep(time.Duration(proccess.Job.IoTime) * time.Second)
		proccess.ScheduleTimes[len(proccess.ScheduleTimes)-1].Complete = time.Now()
		proccess.Job.IoTime = -1
		// submit proccess to cpu to execute
		var scheduleTime = ScheduleTime{
			Submission: time.Now(),
		}
		proccess.ScheduleTimes = append(proccess.ScheduleTimes, scheduleTime)
		// add this job to ready queue
		cpuWorkQueue <- proccess
		log.Println("pid:", proccess.Job.ProcessId, "io-request done")
	}
}
