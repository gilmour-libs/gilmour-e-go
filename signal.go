package gilmour

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func bindSignals(engine *Gilmour) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		engine.Stop()
		log.Println("Shutting down engines.")
		os.Exit(0)
	}()
}
