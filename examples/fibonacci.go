package main

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/gilmour-libs/gilmour-e-go.v4"
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/backends"
)

const (
	fibTopic = "fib"
	FIRST    = 196418
	SECOND   = 317811
)

func makeGilmour() *gilmour.Gilmour {
	redis := backends.MakeRedis("127.0.0.1:6379", "")
	engine := gilmour.Get(redis)
	return engine
}

func bindListeners(e *gilmour.Gilmour) {
	o := gilmour.NewHandlerOpts().SetGroup(fibTopic)

	e.ReplyTo(fibTopic, func(req *gilmour.Request, resp *gilmour.Message) {
		pack := map[string]float64{}
		req.Data(&pack)

		next := pack["first"] + pack["second"]
		fmt.Printf("First %.0f Second %.0f Next %.0f \n", pack["first"], pack["second"], next)
		resp.Send(next)
	}, o)
}

func generator(first, second float64, tick <-chan time.Time, e *gilmour.Gilmour) {
	//Wait for a tick
	<-tick

	packet := map[string]float64{"first": first, "second": second}
	data := gilmour.NewMessage().Send(packet)

	handler := func(req *gilmour.Request, resp *gilmour.Message) {
		if req.Code() != 200 {
			fmt.Println("Error in Handler", req.Code())
			fmt.Println(req.RawData())
			return
		}

		var next float64
		if err := req.Data(&next); err != nil {
			fmt.Println(err)
			return
		}

		generator(second, next, tick, e)
	}

	opts := gilmour.NewRequestOpts().SetHandler(handler)
	e.Request(fibTopic, data, opts)
}

func main() {
	engine := makeGilmour()
	bindListeners(engine)

	engine.Start()

	c := time.Tick(1 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	generator(FIRST, SECOND, c, engine)

	wg.Wait()
	engine.Stop()
}
