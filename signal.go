package gilmour

import (
	"os"
	"os/signal"
	"syscall"
)

func BindSignals(engine *Gilmour) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		engine.Stop()
		log.Info("Shutting down engines.")
		os.Exit(0)
	}()
}
