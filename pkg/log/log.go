package log

import (
	log "github.com/sirupsen/logrus"
)

func Logger() *log.Logger {
	log.SetLevel(log.TraceLevel)
	return log.StandardLogger()
}
