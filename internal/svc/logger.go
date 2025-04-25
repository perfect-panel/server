package svc

import (
	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

func NewLogger(c config.Config) *logger.Logger {
	//log := logger.New(c.Logger)
	//// replace the default logger
	//logger = log
	//return log
	return nil
}
