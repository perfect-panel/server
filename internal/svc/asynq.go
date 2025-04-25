package svc

import (
	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/config"
)

func NewAsynqClient(c config.Config) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: c.Redis.Host, Password: c.Redis.Pass, DB: 5})
}
