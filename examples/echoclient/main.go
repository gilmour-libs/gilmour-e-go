package main

import (
	"fmt"
	"sync"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v4"
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/backends"
)

func echoEngine() *G.Gilmour {
	redis := backends.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(redis)
	return engine
}

func echoRequest(wg *sync.WaitGroup, engine *G.Gilmour, msg string) {
	data := G.NewMessage().SetData(msg)

	handler := func(req *G.Request, resp *G.Message) {
		defer wg.Done()

		var msg string
		if err := req.Data(&msg); err != nil {
			fmt.Println("Echoclient: error", err.Error())
		} else {
			fmt.Println("Echoclient: received", msg)
		}
	}

	opts := G.NewRequestOpts().SetHandler(handler)
	engine.Request("echo", data, opts)
}

func main() {
	engine := echoEngine()
	engine.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go echoRequest(&wg, engine, fmt.Sprintf("Hello: %v", i))
	}

	wg.Wait()
	engine.Stop()
}
