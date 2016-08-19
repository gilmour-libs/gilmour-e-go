package main

import (
	"log"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

type weatherTuple struct {
	Day     int
	Month   string
	Hour    int
	Degrees int
}

func echoEngine() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

func monthLambda(month string) func(m *G.Message) (*G.Message, error) {
	return func(m *G.Message) (*G.Message, error) {
		var data map[string]interface{}
		if err := m.GetData(&data); err != nil {
			panic(err)
		}

		return G.NewMessage().SetData(data[month]), nil
	}
}

func main() {
	e := echoEngine()
	e.Start()

	batch := e.NewPipe(
		e.NewRequest("weather.fetch"),
		e.NewRequest("weather.group"),
		e.NewParallel(
			e.NewPipe(
				e.NewLambda(monthLambda("jan")),
				e.NewParallel(
					e.NewRequest("weather.min"),
					e.NewRequest("weather.max"),
				),
			),
			e.NewPipe(
				e.NewLambda(monthLambda("feb")),
				e.NewParallel(
					e.NewRequest("weather.min"),
					e.NewRequest("weather.max"),
				),
			),
		),
	)

	resp, _ := batch.Execute(G.NewMessage().SetData("pune"))

	var out []*G.Message
	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out = append(out, msg)
	}

	for _, msg := range out {
		monthOut := []*weatherTuple{}
		msg.GetData(&monthOut)

		for _, m := range monthOut {
			log.Println(m)
		}
	}
}
