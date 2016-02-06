package gilmour

import (
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/gilmour-libs/gilmour-e-go.v4/ui"
)

//Call this method with a Gilmour instance for graceful shutdown and cleanup
//on os signals. Dont' know if it works on Windows.
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
		ui.Warn("Shutting down engines.")
		os.Exit(0)
	}()
}
