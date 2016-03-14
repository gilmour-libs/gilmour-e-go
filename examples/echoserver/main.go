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

func echoReply(req *G.Request, resp *G.Message) {
	var msg string
	req.Data(&msg)
	fmt.Println("Echoserver: received", msg)
	resp.SetData(fmt.Sprintf("Pong %v", msg))
}

func bindListeners(g *G.Gilmour) {
	opts := G.NewHandlerOpts().SetGroup("exclusive")
	g.ReplyTo("echo", echoReply, opts)
}

func main() {
	engine := echoEngine()
	bindListeners(engine)

	engine.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
