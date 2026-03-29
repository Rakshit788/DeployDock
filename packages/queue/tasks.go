package tasks

import (
	"github.com/hibiken/asynq"
)

var queue *asynq.Client

func InitQueue() {
	queue = asynq.NewClient(asynq.RedisClientOpt{
		Addr: "redis:6379",
	})
}
