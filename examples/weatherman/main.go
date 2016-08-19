package main

import (
	"sort"
	"sync"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

func echoEngine() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

func glog(g *G.Gilmour, msg string) {
	g.Signal("example.log", G.NewMessage().SetData(msg))
}

//Fetch a remote file from the URL received in Request.
func fetchReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		glog(g, "Fetching weather data")

		var city string
		if err := req.Data(&city); err != nil {
			panic(err)
		}

		//Send back the contents.
		resp.SetData(weatherData(city))
	}
}

func groupReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		glog(g, "Grouping weather data to respective months")

		var data []weatherTuple
		if err := req.Data(&data); err != nil {
			panic(err)
		}

		mapped := make(map[string][]weatherTuple)
		for _, tup := range data {
			mapped[tup.Month] = append(mapped[tup.Month], tup)
		}

		//Send back the contents
		resp.SetData(mapped)
	}
}

type By func(p1, p2 *weatherTuple) bool

func (by By) Sort(wt []weatherTuple) {
	ps := &weatherSorter{
		weathers: wt,
		by:       by,
	}
	sort.Sort(ps)
}

type weatherSorter struct {
	weathers []weatherTuple
	by       func(p1, p2 *weatherTuple) bool
}

func (s *weatherSorter) Len() int {
	return len(s.weathers)
}

func (s *weatherSorter) Swap(i, j int) {
	s.weathers[i], s.weathers[j] = s.weathers[j], s.weathers[i]
}

func (s *weatherSorter) Less(i, j int) bool {
	return s.by(&s.weathers[i], &s.weathers[j])
}

func minReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		glog(g, "Computing minimum temperature")

		var data []weatherTuple
		if err := req.Data(&data); err != nil {
			panic(err)
		}

		degrees := func(p1, p2 *weatherTuple) bool {
			return p1.Degrees < p2.Degrees
		}

		By(degrees).Sort(data)
		resp.SetData(data[0])
	}
}

func maxReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		glog(g, "Computing maximum temperature")

		var data []weatherTuple
		if err := req.Data(&data); err != nil {
			panic(err)
		}

		degrees := func(p1, p2 *weatherTuple) bool {
			return p1.Degrees > p2.Degrees
		}

		By(degrees).Sort(data)
		resp.SetData(data[0])
	}
}

//Bind all service endpoints to their topics.
func bindListeners(g *G.Gilmour) {
	//Third argument is for Opts, which in this case is nil
	g.ReplyTo("weather.fetch", fetchReply(g), nil)
	g.ReplyTo("weather.group", groupReply(g), nil)
	g.ReplyTo("weather.min", minReply(g), nil)
	g.ReplyTo("weather.max", maxReply(g), nil)
}

func main() {
	engine := echoEngine()
	bindListeners(engine)

	engine.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
