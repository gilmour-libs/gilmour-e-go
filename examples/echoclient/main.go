package main

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/gilmour-libs/gilmour-e-go.v4"
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/backends"
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/ui"
)

const echoTopic = "echo"

func echoEngine() *gilmour.Gilmour {
	redis := backends.MakeRedis("127.0.0.1:6379", "")
	engine := gilmour.Get(redis)
	return engine
}

func echoRequest(wg *sync.WaitGroup, engine *gilmour.Gilmour, msg string) {
	data := gilmour.NewMessage().Send(msg)

	handler := func(req *gilmour.Request, resp *gilmour.Message) {
		defer wg.Done()

		var msg string
		if err := req.Data(&msg); err != nil {
			fmt.Println("Echoclient: error", err.Error())
		} else {
			fmt.Println("Echoclient: received", msg)
		}
	}

	opts := gilmour.NewRequestOpts().SetHandler(handler)
	engine.Request(echoTopic, data, opts)
}

func main() {
	ui.SetLevel(ui.Levels.Message)
	engine := echoEngine()
	engine.Start()

	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		echoRequest(&wg, engine, fmt.Sprintf("Hello: %v", i))
		time.Sleep(2 * time.Second)
	}

	wg.Wait()
	engine.Stop()
}
