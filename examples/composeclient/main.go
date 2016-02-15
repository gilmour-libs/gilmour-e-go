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

type MyComposer struct {
}

func (m *MyComposer) Execute(msg *G.Message) <-chan *G.Message {
	out := []*G.Message{}
	msg.GetData(&out)

	pWords := make([][]string, 3)

	for _, o := range out {
		popular := struct {
			WordLength int
			Score      int
			Words      []string
		}{}
		o.GetData(&popular)
		pWords[popular.WordLength-3] = popular.Words
	}

	outChan := make(chan *G.Message, 1)
	outChan <- G.NewMessage().SetData(pWords)
	close(outChan)
	return outChan
}

func (m *MyComposer) IsStreaming() bool {
	return false
}

func main() {
	engine := echoEngine()
	engine.Start()

	batch := engine.NewPipe(
		engine.NewRequestComposition("example.fetch"),
		engine.NewRequestComposition("example.words"),
		engine.NewRequestComposition("example.stopfilter"),
		engine.NewRequestComposition("example.count"),
		engine.NewParallel(
			engine.NewRequestComposition("example.popular3"),
			engine.NewRequestComposition("example.popular4"),
			engine.NewRequestComposition("example.popular5"),
		),
		&MyComposer{},
	)

	data := G.NewMessage()
	data.SetData("https://s3-us-west-1.amazonaws.com/ds-data-sample/test.txt")

	msg := <-batch.Execute(data)
	expected := [][]string{}
	if err := msg.GetData(&expected); err != nil {
		log.Println(err)
	} else {
		for ix, words := range expected {
			log.Printf("Popular %v letter words: %v\n", ix+3, words)
		}
	}
}
