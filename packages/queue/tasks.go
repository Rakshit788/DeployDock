package tasks

import (
	"github.com/hibiken/asynq"
)

var Queue *asynq.Client

func InitQueue(addr string) {
	if addr == "" {
		addr = "redis:6379"
	}

	Queue = asynq.NewClient(asynq.RedisClientOpt{
		Addr: addr,
	})
}
