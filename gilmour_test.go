package gilmour

import (
	redigo "github.com/garyburd/redigo/redis"
	"gopkg.in/gilmour-libs/gilmour-e-go.v0"
	"gopkg.in/gilmour-libs/gilmour-e-go.v0/backends"
	"gopkg.in/gilmour-libs/gilmour-e-go.v0/logger"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	PingTopic    = "ping"
	PingResponse = "pong"
	SleepTopic   = "sleepy-ping"
)

var engine *gilmour.Gilmour
var redis *backends.Redis

func compare(X, Y []string) []string {
	m := make(map[string]int)

	for _, y := range Y {
		m[y]++
	}

	var ret []string
	for _, x := range X {
		if m[x] > 0 {
			m[x]--
			continue
		}
		ret = append(ret, x)
	}

	return ret
}

func TestSubscribePing(t *testing.T) {
	timeout := 3
	handler_opts := gilmour.MakeHandlerOpts().SetTimeout(timeout)
	sub := engine.Subscribe(PingTopic, func(req *gilmour.Request, resp *gilmour.Response) {
		var x string
		req.Data(&x)
		req.Logger.Debug(PingTopic, "Received", x)
		resp.Respond(PingResponse)
	}, handler_opts)

	actualTimeout := sub.GetOpts().GetTimeout()

	if actualTimeout != timeout {
		t.Error("Handler should have timeout of", timeout, "seconds. Found", actualTimeout)
	}
}

func TestSubscribeSleep(t *testing.T) {
	sub := engine.Subscribe(
		SleepTopic,
		func(req *gilmour.Request, resp *gilmour.Response) {
			var delay int
			req.Data(&delay)
			time.Sleep(time.Duration(delay) * time.Second)
			resp.Respond(PingResponse)
		}, nil,
	)

	actualTimeout := sub.GetOpts().GetTimeout()

	if actualTimeout != 600 {
		t.Error("Handler should have default timeout of 600, Found", actualTimeout)
	}
}

func TestHealthGetAll(t *testing.T) {
	conn := redis.GetConn()
	defer conn.Close()

	idents, err := redigo.StringMap(conn.Do("HGETALL", redis.GetHealthIdent()))
	if err != nil {
		t.Error(err)
	}

	val, ok := idents[engine.GetIdent()]
	if !ok || val != "true" {
		t.Error("Ident is missing in the Health ident")
	}
}

func TestHealthResponse(t *testing.T) {
	out_chan := make(chan string, 1)

	opts := gilmour.NewPublisher().
		SetData("is-healthy?").
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {
		x := []string{}
		expected := []string{PingTopic, SleepTopic}

		req.Data(&x)
		skew := compare(expected, x)

		if len(skew) == 0 {
			out_chan <- "healthy"
		} else {
			out_chan <- "false"
		}
	})

	_, err := engine.Publish(redis.HealthTopic(engine.GetIdent()), opts)
	if err != nil {
		t.Error(err)
	}

	select {
	case result := <-out_chan:
		if result != "healthy" {
			t.Error("Response should be healthy. Found", result)
		}
	case <-time.After(time.Second * 5):
		t.Error("Response should be", PingResponse, "timed out instead")
	}
}

func TestSendAndReceive(t *testing.T) {
	out_chan := make(chan string, 1)

	opts := gilmour.NewPublisher().
		SetData("ping?").
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {
		var x string
		req.Data(&x)
		out_chan <- x
	})

	_, err := engine.Publish(PingTopic, opts)
	if err != nil {
		t.Error(err)
	}

	select {
	case result := <-out_chan:
		if result != PingResponse {
			t.Error("Response should be", PingResponse, "Found", result)
		}
	case <-time.After(time.Second * 5):
		t.Error("Response should be", PingResponse, "timed out instead")
	}
}

func TestWithTimeout(t *testing.T) {
	out_chan := make(chan string, 1)

	opts := gilmour.NewPublisher().
		SetData(5).    //Will Sleep for 5 seconds.
		SetTimeout(2). //Will expect response in 2 seconds.
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {
		var x string
		req.Data(&x)
		out_chan <- x
	})

	_, err := engine.Publish(SleepTopic, opts)
	if err != nil {
		t.Error(err)
	}

	select {
	case result := <-out_chan:
		if result != "Execution timed out" {
			t.Error("Response should be", PingResponse, "Found", result)
		}
	case <-time.After(time.Second * 5):
		t.Error("Response should be", PingResponse, "timed out instead")
	}
}

func TestSansListener(t *testing.T) {
	opts := gilmour.NewPublisher().SetData("ping?")
	_, err := engine.Publish("ping-sans-listener", opts)
	if err != nil {
		t.Error(err)
	}
}

func TestConfirmSansListener(t *testing.T) {
	pub_opts := gilmour.NewPublisher().
		SetData("ping?").
		ConfirmSubscriber()

	_, err := engine.Publish("ping-confirm-sans-listener", pub_opts)
	if err != nil {
		t.Error(err)
	}
}

func TestHandlerConfirmSansListener(t *testing.T) {
	out_chan := make(chan int, 1)

	pub_opts := gilmour.NewPublisher().
		SetData("ping?").
		ConfirmSubscriber().
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {
		out_chan <- req.Code()
	})

	_, err := engine.Publish("ping-confirm-sans-listener", pub_opts)
	if err != nil {
		out_chan <- 500
	}

	select {
	case code := <-out_chan:
		if code != 404 {
			t.Error("Response should be", 404, "Found", code)
		}
	case <-time.After(time.Second * 5):
		t.Error("Response should be", 404, "timed out instead")
	}
}

func waitBeforeExiting(interval int) {
	var wg sync.WaitGroup
	wg.Add(1)

	time.AfterFunc(time.Duration(interval)*time.Second, func() {
		wg.Done()
	})

	wg.Wait()
}

func TestMain(m *testing.M) {
	redis = backends.MakeRedis("127.0.0.1:6379")
	engine = gilmour.Get(redis)

	engine.SetHealthCheckEnabled()
	engine.Start()

	logger.Logger.Info("Starting Engine")

	status := m.Run()

	waitBeforeExiting(2) //Wait 5 seconds before exiting.
	engine.Stop()
	waitBeforeExiting(2) //Wait 2 more seconds before exiting.

	os.Exit(status)
}
