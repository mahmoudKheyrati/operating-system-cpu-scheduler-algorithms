package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

type SchedulerConfig struct {
	Port                                     int
	RoundRobinTimeQuantum                    int
	MultilevelFeedbackQueueLevelsTimeQuantum []int
}

var once sync.Once
var config *SchedulerConfig

func GetSchedulerConfig() *SchedulerConfig {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalln(err)
		}
		config = &SchedulerConfig{}
		config.Port = viper.GetInt("port")
		config.RoundRobinTimeQuantum = viper.GetInt("scheduler.round_robin.time_quantum")
		config.MultilevelFeedbackQueueLevelsTimeQuantum = viper.GetIntSlice("scheduler.multilevel_feedback_queue.levels_time_quantum")

	})

	return config
}
