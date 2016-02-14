package main

import (
	"log"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v4"
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/backends"
)

func echoEngine() *G.Gilmour {
	redis := backends.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(redis)
	return engine
}

func main() {
	engine := echoEngine()
	engine.Start()

	batch := engine.NewPipe(
		engine.NewRequestComposition("example.fetch"),
		engine.NewRequestComposition("example.count"),
		engine.NewParallel(
			engine.NewRequestComposition("example.popular3"),
			engine.NewRequestComposition("example.popular4"),
			engine.NewRequestComposition("example.popular5"),
		),
	)

	data := G.NewMessage()
	data.Send("https://s3-us-west-1.amazonaws.com/ds-data-sample/test.txt")

	out := []*G.Message{}
	(<-batch.Execute(data)).Receive(&out)

	for _, msg := range out {
		popular := struct {
			WordLength int
			Score      int
			Words      []string
		}{}
		msg.Receive(&popular)
		log.Println(popular)
	}
}
