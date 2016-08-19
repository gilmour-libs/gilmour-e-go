package main

import (
	"log"
	"sync"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

func echoEngine() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	engine := echoEngine()
	engine.Slot("example.log", func(req *G.Request) {
		var msg string
		if err := req.Data(&msg); err != nil {
			log.Println("Cannot parse log %v", err.Error())
			return
		}

		log.Println(req.Sender(), "->", msg)
	}, nil)

	engine.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
