package util

import "os-project/internal/responses"

func CalculateAverage(proccessDetails []responses.ProcessResponse) (averageWaitingTime, averageResponseTime, averageTimeAroundTime float64) {
	var waitingTimeSum float64
	var responseTimeSum float64
	var turnAroundTimeSum float64

	for _, proccess := range proccessDetails {
		waitingTimeSum += proccess.WaitingTime
		responseTimeSum += proccess.ResponseTime
		turnAroundTimeSum += proccess.TurnAroundTime
	}

	proccessCount := float64(len(proccessDetails))

	averageWaitingTime = waitingTimeSum / proccessCount
	averageResponseTime = responseTimeSum / proccessCount
	averageTimeAroundTime = turnAroundTimeSum / proccessCount
	return
}
