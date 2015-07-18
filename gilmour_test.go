package gilmour

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"gopkg.in/gilmour-libs/gilmour-e-go.v0"
	"gopkg.in/gilmour-libs/gilmour-e-go.v0/backends"
	"gopkg.in/gilmour-libs/gilmour-e-go.v0/logger"
	"math/rand"
	"os"
	"strings"
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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

func isTopicSubscribed(topic string) (has bool, err error) {
	conn := redis.GetConn()
	defer conn.Close()

	idents, err2 := redigo.Strings(conn.Do("PUBSUB", "CHANNELS"))
	if err2 != nil {
		err = err2
		return
	}

	for _, t := range idents {
		if t == topic {
			has = true
			break
		}
	}

	return
}

func TestHealthSubscribe(t *testing.T) {
	engine.SetHealthCheckEnabled()

	topic := redis.HealthTopic(engine.GetIdent())
	if has, _ := isTopicSubscribed(topic); !has {
		t.Error(topic, "should have been subscribed")
	}
}

func TestSubscribePing(t *testing.T) {
	timeout := 3
	handler_opts := gilmour.MakeHandlerOpts().SetTimeout(timeout)
	sub, _ := engine.Subscribe(PingTopic, func(req *gilmour.Request, resp *gilmour.Response) {
		var x string
		req.Data(&x)
		req.Logger.Debug(PingTopic, "Received", x)
		resp.Respond(PingResponse)
	}, handler_opts)

	actualTimeout := sub.GetOpts().GetTimeout()

	if actualTimeout != timeout {
		t.Error("Handler should have timeout of", timeout, "seconds. Found", actualTimeout)
	}

	if has, _ := isTopicSubscribed(PingTopic); !has {
		t.Error("Topic", PingTopic, "should have been subscribed")
	}
}

func TestWildcardGroup(t *testing.T) {
	opts := gilmour.MakeHandlerOpts().SetGroup("wildcard_group")
	topic := fmt.Sprintf("%v*", PingTopic)
	_, err := engine.Subscribe(topic, func(req *gilmour.Request, resp *gilmour.Response) {}, opts)

	if err == nil || !strings.Contains(err.Error(), "cannot have") {
		t.Error("Wildcars cannot belong to a Group")
	}
}

func TestSubscribeSleep(t *testing.T) {
	sub, _ := engine.Subscribe(
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

	if has, _ := isTopicSubscribed(SleepTopic); !has {
		t.Error("Topic", SleepTopic, "should have been subscribed")
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

func TestUnsubscribe(t *testing.T) {
	topic := randSeq(10)
	sub, _ := engine.Subscribe(topic, func(req *gilmour.Request, resp *gilmour.Response) {}, nil)

	if has, _ := isTopicSubscribed(topic); !has {
		t.Error(topic, "should have been subscribed")
	}

	engine.Unsubscribe(topic, sub)

	if has, _ := isTopicSubscribed(topic); has {
		t.Error("Topic", topic, "should have been unsubscribed")
	}
}

func TestRelayFromSubscriber(t *testing.T) {
}

func TestSendOnceReceiveTwice(t *testing.T) {
	topic := randSeq(10)
	count := 2

	out := []string{}
	subs := []*gilmour.Subscription{}
	out_chan := make(chan string, count)

	// Subscribe x no. of times
	for i := 0; i < count; i++ {
		data := fmt.Sprintf("hello %v", i)
		sub, _ := engine.Subscribe(topic,
			func(_ *gilmour.Request, _ *gilmour.Response) { out_chan <- data },
			nil)
		subs = append(subs, sub)
	}

	//Publish a message to random topic
	pub_opts := gilmour.NewPublisher().SetData("ping?")
	engine.Publish(topic, pub_opts)

	// Select Loop over channel, and timeout eventually.
	for i := 0; i < count; i++ {
		select {
		case result := <-out_chan:
			out = append(out, result)
		case <-time.After(time.Second * 5):
			t.Error("Response should be twice, timed out instead")
		}
	}

	// Results should be received twice.
	if len(out) != count {
		t.Error("Response should be returned ", count, "items. Found", out)
	}

	// cleanup. Ubsubscribe from all subscribed channels.
	for _, sub := range subs {
		engine.Unsubscribe(topic, sub)
	}

	// Confirm that the topic was Indeed unsubscribed.
	if has, _ := isTopicSubscribed(topic); has {
		t.Error("Topic", topic, "should have been unsubscribed")
	}
}

func TestSendOnceReceiveOnce(t *testing.T) {
	topic := randSeq(10)
	count := 2

	out := []string{}
	subs := []*gilmour.Subscription{}
	out_chan := make(chan string, count)

	// Subscribe x no. of times
	for i := 0; i < count; i++ {
		data := fmt.Sprintf("hello %v", i)
		opts := gilmour.MakeHandlerOpts().SetGroup("unique")
		sub, _ := engine.Subscribe(topic,
			func(_ *gilmour.Request, _ *gilmour.Response) { out_chan <- data },
			opts)
		subs = append(subs, sub)
	}

	//Publish a message to random topic
	pub_opts := gilmour.NewPublisher().SetData("ping?")
	engine.Publish(topic, pub_opts)

	// Select Case, once and that should work.
	select {
	case result := <-out_chan:
		out = append(out, result)
	case <-time.After(time.Second * 2):
		t.Error("Response should be twice, timed out instead")
	}

	//Same code, should fail the second time.
	select {
	case <-out_chan:
		t.Error("Should not receive a value on second handler.")
	case <-time.After(time.Second * 2):
		out = append(out, "Ok")
	}

	// Results should be received twice.
	if len(out) != count {
		t.Error("Response should be returned ", count, "items. Found", out)
	}

	// cleanup. Ubsubscribe from all subscribed channels.
	for _, sub := range subs {
		engine.Unsubscribe(topic, sub)
	}

	// Confirm that the topic was Indeed unsubscribed.
	if has, _ := isTopicSubscribed(topic); has {
		t.Error("Topic", topic, "should have been unsubscribed")
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

func TestReceiveOnWildcard(t *testing.T) {
	topic := fmt.Sprintf("%v*", PingTopic)
	out_chan := make(chan string, 1)

	//Subscribe to the wildcard topic.
	sub, _ := engine.Subscribe(
		topic,
		func(_ *gilmour.Request, _ *gilmour.Response) {
			out_chan <- PingResponse
		},
		nil,
	)

	//Publish a message to random topic
	pub_opts := gilmour.NewPublisher().SetData("ping?")
	engine.Publish("pingworld", pub_opts)

	// Select Case, once and that should work.
	select {
	case <-out_chan:
		// True Case.
	case <-time.After(time.Second * 2):
		t.Error("Response should be received, timed out instead")
	}

	// cleanup. Ubsubscribe from all subscribed channels.
	engine.Unsubscribe(topic, sub)

	// Confirm that the topic was Indeed unsubscribed.
	if has, _ := isTopicSubscribed(topic); has {
		t.Error("Topic", topic, "should have been unsubscribed")
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

func TestPublisherTimeout(t *testing.T) {
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

func TestNoPublisher(t *testing.T) {
	_, err := engine.Publish(PingTopic, nil)
	if err == nil || !strings.Contains(err.Error(), "provide publisher") {
		t.Error("Must provide publisher to be published")
	}
}

func TestSubscriberTimeout(t *testing.T) {
	out_chan := make(chan string, 1)
	topic := "sleep_delayed"

	sub, _ := engine.Subscribe(
		topic,
		func(req *gilmour.Request, resp *gilmour.Response) {
			time.Sleep(time.Second * 4)
			resp.Respond(PingResponse)
		},
		gilmour.MakeHandlerOpts().SetTimeout(3),
	)

	opts := gilmour.NewPublisher().
		SetData("send"). //Will Sleep for 5 seconds.
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {
		var x string
		req.Data(&x)
		out_chan <- x
	})

	engine.Publish(topic, opts)

	select {
	case result := <-out_chan:
		if result != "Execution timed out" {
			t.Error("Response should be", PingResponse, "Found", result)
		}
	case <-time.After(time.Second * 5):
		t.Error("Response should be", PingResponse, "timed out instead")
	}

	engine.Unsubscribe(topic, sub)
}

func TestHandlerException(t *testing.T) {
	out_chan := make(chan int, 1)
	topic := randSeq(10)

	sub, _ := engine.Subscribe(
		topic,
		func(req *gilmour.Request, resp *gilmour.Response) {
			// Just to induce errors, access a null pointers's method.
			var x *HandlerOpts
			log.Debug(x.GetGroup())
		}, nil,
	)

	opts := gilmour.NewPublisher().
		SetData("send"). //Will Sleep for 5 seconds.
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {
		out_chan <- req.Code()
	})

	engine.Publish(topic, opts)

	select {
	case result := <-out_chan:
		if result != 500 {
			t.Error("Response should raise exception")
		}
	case <-time.After(time.Second * 5):
		t.Error("Response should have raised Exception, timed out instead")
	}

	engine.Unsubscribe(topic, sub)
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
		if !strings.Contains(err.Error(), "active subscribers") {
			t.Error(err.Error())
		}
	}
}

func TestHandlerConfirmSansListener(t *testing.T) {
	pub_opts := gilmour.NewPublisher().
		SetData("ping?").
		ConfirmSubscriber().
		SetHandler(func(req *gilmour.Request, resp *gilmour.Response) {})

	_, err := engine.Publish("ping-confirm-sans-listener", pub_opts)
	if err != nil {
		if !strings.Contains(err.Error(), "active subscribers") {
			t.Error(err.Error())
		}
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

	engine.Start()

	logger.Logger.Info("Starting Engine")

	status := m.Run()

	waitBeforeExiting(2) //Wait 5 seconds before exiting.
	engine.Stop()
	waitBeforeExiting(2) //Wait 2 more seconds before exiting.

	os.Exit(status)
}
