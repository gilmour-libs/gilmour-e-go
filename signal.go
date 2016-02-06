package gilmour

import (
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/gilmour-libs/gilmour-e-go.v4/ui"
)

// Call to start listening on signals for graceful shutdown & cleanup. By
// default Gilmour listens to SIGHUP, SIGINT, SIGTERM, SIGQUIT, however
// you can override this by providing a list of desired signals instead.
func (g *Gilmour) BindSignals(signals ...os.Signal) {
	sigc := make(chan os.Signal, 1)

	if len(signals) == 0 {
		signals = []os.Signal{
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		}
	}

	signal.Notify(sigc, signals...)

	go func() {
		<-sigc
		ui.Warn("Shutting down engines.")
		g.Stop()
	}()
}
