package main

import (
	"fmt"
	"sync"

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

func echoReply(req *gilmour.Request, resp *gilmour.Message) {
	var msg string
	req.Data(&msg)
	fmt.Println("Echoserver: received", msg)
	resp.Send(fmt.Sprintf("Pong %v", msg))
}

func bindListeners(g *gilmour.Gilmour) {
	opts := gilmour.NewHandlerOpts().SetGroup(echoTopic)
	g.ReplyTo(echoTopic, echoReply, opts)
}

func main() {
	ui.SetLevel(ui.Levels.Message)
	engine := echoEngine()
	bindListeners(engine)

	engine.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
