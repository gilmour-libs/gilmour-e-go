package logger

import (
	log "gopkg.in/inconshreveable/log15.v2"
)

var Logger = log.New()

func init() {
	Logger.SetHandler(log.StderrHandler)
}

func Module(module string) log.Logger {
	return Logger.New("module", module)
}

func Handler(sender string) log.Logger {
	return Logger.New("sender", sender)
}
