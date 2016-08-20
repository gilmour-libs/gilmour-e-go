package main

import (
	"log"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

const url = "https://s3-us-west-1.amazonaws.com/ds-data-sample/test.txt"

func echoEngine() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

type popular struct {
	WordLength int
	Score      int
	Words      []string
}

func converge(msg *G.Message) (*G.Message, error) {
	out := []*popular{}
	msg.GetData(&out)

	pWords := make([][]string, 3)

	for _, o := range out {
		pWords[o.WordLength-3] = o.Words
	}

	return G.NewMessage().SetData(pWords), nil
}

func main() {
	engine := echoEngine()
	engine.Start()

	pipe := engine.NewPipe(
		engine.NewRequest("example.fetch"),
		engine.NewRequest("example.words"),
		engine.NewRequest("example.stopfilter"),
		engine.NewRequest("example.count"),
		engine.NewParallel(
			engine.NewRequest("example.popular3"),
			engine.NewRequest("example.popular4"),
			engine.NewRequest("example.popular5"),
		),
		engine.NewLambda(converge),
	)

	resp, err := pipe.Execute(G.NewMessage().SetData(url))
	if err != nil {
		log.Println(err)
		return
	}

	expected := [][]string{}
	if err := resp.Next().GetData(&expected); err != nil {
		log.Println(err)
	} else {
		for _, words := range expected {
			log.Printf("Popular words: %v\n", words)
		}
	}
}
