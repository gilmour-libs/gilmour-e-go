package main

import (
	"fmt"
	"sync"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

func echoEngine() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

func echoRequest(wg *sync.WaitGroup, engine *G.Gilmour, msg string) {
	req := engine.NewRequest("echo")
	resp, err := req.Execute(G.NewMessage().SetData(msg))
	if err != nil {
		fmt.Println("Echoclient: error", err.Error())
	}

	defer wg.Done()

	var output string
	if err := resp.Next().GetData(&output); err != nil {
		fmt.Println("Echoclient: error", err.Error())
	} else {
		fmt.Println("Echoclient: received", output)
	}
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
