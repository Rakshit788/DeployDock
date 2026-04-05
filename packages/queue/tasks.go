package tasks

import (
	"github.com/hibiken/asynq"
)

var Queue *asynq.Client

func InitQueue() {
	Queue = asynq.NewClient(asynq.RedisClientOpt{
		Addr: "redis:6379",
	})
}
