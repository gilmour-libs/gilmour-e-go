package main

import (
	"log"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v4"
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/backends"
)

type weatherTuple struct {
	Day     int
	Month   string
	Hour    int
	Degrees int
}

func echoEngine() *G.Gilmour {
	redis := backends.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(redis)
	return engine
}

func monthLambda(month string) func(m *G.Message) *G.Message {
	return func(m *G.Message) *G.Message {
		var data map[string]interface{}
		if err := m.GetData(&data); err != nil {
			panic(err)
		}

		return G.NewMessage().SetData(data[month])
	}
}

func main() {
	engine := echoEngine()
	engine.Start()

	batch := engine.NewPipe(
		engine.NewRequestComposition("weather.fetch"),
		engine.NewRequestComposition("weather.group"),
		engine.NewParallel(
			engine.NewPipe(
				engine.NewFuncComposition(monthLambda("jan")),
				engine.NewParallel(
					engine.NewRequestComposition("weather.min"),
					engine.NewRequestComposition("weather.max"),
				),
			),
			engine.NewPipe(
				engine.NewFuncComposition(monthLambda("feb")),
				engine.NewParallel(
					engine.NewRequestComposition("weather.min"),
					engine.NewRequestComposition("weather.max"),
				),
			),
		),
	)

	data := G.NewMessage()
	data.SetData("pune")

	var out []*G.Message

	(<-batch.Execute(data)).GetData(&out)

	for _, msg := range out {
		var monthOut []*G.Message
		msg.GetData(&monthOut)

		for _, m := range monthOut {
			var tup weatherTuple
			m.GetData(&tup)
			log.Printf(
				"Temperature on %v/%v at %v00 hours was %vºF",
				tup.Day, tup.Month, tup.Hour, tup.Degrees,
			)
		}
	}
}
