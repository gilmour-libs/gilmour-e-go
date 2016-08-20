package main

import (
	"fmt"
	"sync"
	"time"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

const (
	fibTopic = "fib"
	FIRST    = 0
	SECOND   = 1
)

func makeGilmour() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

func fibRequest(req *G.Request, resp *G.Message) {
	pack := map[string]float64{}
	req.Data(&pack)

	next := pack["first"] + pack["second"]
	fmt.Printf("First %.0f Second %.0f Next %.0f \n",
		pack["first"], pack["second"], next)
	resp.SetData(next)
}

func bindListeners(e *G.Gilmour) {
	o := G.NewHandlerOpts().SetGroup(fibTopic)
	e.ReplyTo(fibTopic, fibRequest, o)
}

func generator(first, second float64, tick <-chan time.Time, e *G.Gilmour) {
	//Wait for a tick
	<-tick

	packet := map[string]float64{"first": first, "second": second}
	data := G.NewMessage().SetData(packet)

	req := e.NewRequest(fibTopic)
	resp, err := req.Execute(data)
	if err != nil {
		fmt.Println("Error in Handler", err)
		return
	}

	if resp.Code() != 200 {
		fmt.Println("Error in Handler", resp.Code())
		return
	}

	var next float64
	msg := resp.Next()
	if err := msg.GetData(&next); err != nil {
		fmt.Println(err)
		return
	}

	generator(second, next, tick, e)
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
