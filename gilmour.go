package gilmour

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/gilmour-libs/gilmour-e-go.v1/protocol"
	"gopkg.in/gilmour-libs/gilmour-e-go.v1/ui"
)

func Get(backend Backend) *Gilmour {
	x := Gilmour{}
	x.subscriber = NewSubscriptionManager()
	x.addBackend(backend)
	return &x
}

type Gilmour struct {
	enableHealthCheck bool
	identMutex        sync.RWMutex
	subscriberMutex   sync.RWMutex
	backend           Backend
	ident             string
	errorMethod       string
	subscriber        Subscriber
}

func (self *Gilmour) Start() {
	sink := make(chan *protocol.Message)
	self.backend.Start(sink)
	go self.keepListening(sink)
}

func (self *Gilmour) Stop() {
	defer self.unregisterIdent()
	defer self.backend.Stop()

	for topic, handlers := range self.getAllSubscribers() {
		for _, h := range handlers {
			self.UnsubscribeSlot(topic, h)
			self.UnsubscribeReply(topic, h)
		}
	}
}

func (self *Gilmour) keepListening(sink <-chan *protocol.Message) {
	for {
		msg := <-sink
		go self.processMessage(msg)
	}
}

func (self *Gilmour) processMessage(msg *protocol.Message) {
	subs, ok := self.getSubscribers(msg.Key)
	if !ok || len(subs) == 0 {
		ui.Warn("Message cannot be processed. No subs found for key %v", msg.Key)
		return
	}

	for _, s := range subs {

		if s.GetOpts() != nil && s.GetOpts().IsOneShot() {
			ui.Message("Unsubscribing one shot response topic %v", msg.Topic)
			go self.UnsubscribeReply(msg.Key, s)
		}

		self.executeSubscriber(s, msg.Topic, msg.Data)
	}
}

func (self *Gilmour) executeSubscriber(s *Subscription, topic string, data interface{}) {
	m, err := ParseMessage(data)
	if err != nil {
		ui.Alert(err.Error())
		return
	}

	opts := s.GetOpts()
	if opts.GetGroup() != protocol.BLANK {
		if !self.backend.AcquireGroupLock(opts.GetGroup(), m.GetSender()) {
			ui.Warn(
				"Unable to acquire Lock. Topic %v Group %v Sender %v",
				topic, opts.GetGroup(), m.GetSender(),
			)
			return
		}
	}

	go self.handleRequest(s, topic, m)
}

func (self *Gilmour) handleRequest(s *Subscription, topic string, m *Message) {
	senderId := m.GetSender()

	req := NewRequest(topic, m)

	res := &Message{}
	res.SetSender(self.backend.ResponseTopic(senderId))

	done := make(chan bool, 1)

	//Executing Request
	go func(done chan<- bool) {

		// Schedule a function to recover in case handler runs into an error.
		// Read more: https://gist.github.com/meson10/d56eface6f87c664d07d

		defer func() {
			err := recover()
			if err == nil {
				return
			}

			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			buffer := string(buf)
			res.SetData(buffer).SetCode(500)

			done <- true

		}()

		s.GetHandler()(req, res)
		done <- true
	}(done)

	// Start a timeout handler, which writes on the Done channel, ahead of the
	// Handler. This might result in a RACE condition, as there is no way to
	// kill a goroutine, since they are not preemptive.

	timeout := s.GetOpts().GetTimeout()
	time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		done <- false
	})

	status := <-done

	if !s.GetOpts().IsSlot() {
		if status == false {
			self.sendTimeout(senderId, res.GetSender())
		} else {
			if res.GetCode() == 0 {
				res.SetCode(200)
			}

			if err := self.publish(res.GetSender(), res); err != nil {
				ui.Alert(err.Error())
			}
		}

	} else if status == false {
		// Inform the error catcher, If there is no handler for this Request
		// but the request had failed. This is automatically handled in case
		// of a response being written via Publisher.
		request := string(req.StringData())
		errMsg := protocol.MakeError(499, topic, request, "", req.Sender(), "")
		self.ReportError(errMsg)
	}
}

func (self *Gilmour) sendTimeout(senderId, channel string) {
	msg := &Message{}
	msg.SetSender(senderId).SetCode(499).SetData("Execution timed out")
	if err := self.publish(channel, msg); err != nil {
		ui.Alert(err.Error())
	}
}

func (self *Gilmour) addBackend(backend Backend) {
	self.backend = backend
}

func (self *Gilmour) GetIdent() string {
	self.identMutex.Lock()
	defer self.identMutex.Unlock()

	if self.ident == protocol.BLANK {
		self.ident = protocol.MakeIdent()
	}

	return self.ident
}

func (self *Gilmour) registerIdent() {
	ident := self.GetIdent()
	err := self.backend.RegisterIdent(ident)
	if err != nil {
		panic(err)
	}
}

func (self *Gilmour) unregisterIdent() {
	if !self.IsHealthCheckEnabled() {
		return
	}

	ident := self.GetIdent()
	err := self.backend.UnregisterIdent(ident)
	if err != nil {
		panic(err)
	}
}

func (self *Gilmour) IsHealthCheckEnabled() bool {
	return self.enableHealthCheck
}

func (self *Gilmour) SetHealthCheckEnabled() *Gilmour {
	self.enableHealthCheck = true
	subscribeHealth(self)
	self.registerIdent()
	return self
}

func (self *Gilmour) getAllSubscribers() map[string][]*Subscription {
	return self.subscriber.GetAll()
}

func (self *Gilmour) getSubscribers(topic string) ([]*Subscription, bool) {
	return self.subscriber.Get(topic)
}

func (self *Gilmour) removeSubscribers(topic string) {
	self.subscriber.DeleteAll(topic)
}

func (self *Gilmour) removeSubscriber(topic string, s *Subscription) {
	self.subscriber.Delete(topic, s)
}

func (self *Gilmour) addSubscriber(t string, h Handler, o *HandlerOpts) *Subscription {
	return self.subscriber.Add(t, h, o)
}

func isDuplicateExclusive(topic, group string) bool {
	return false
}

func (self *Gilmour) isExclusiveDuplicate(topic, group string) bool {
	subs, ok := self.getSubscribers(topic)
	if ok {
		for _, s := range subs {
			if s.GetOpts().GetGroup() == group {
				return true
			}
		}
	}

	return false
}

func (self *Gilmour) ReplyTo(topic string, h Handler, opts *HandlerOpts) (*Subscription, error) {
	if strings.Contains(topic, "*") {
		return nil, errors.New("ReplyTo cannot have wildcard topics")
	}

	if opts == nil {
		opts = &HandlerOpts{}
	}

	if opts.GetGroup() == "" {
		opts.SetGroup("_default")
	}

	return self.subscribe(self.requestDestination(topic), h, opts)
}

func (self *Gilmour) Slot(topic string, h Handler, opts *HandlerOpts) (*Subscription, error) {
	if opts == nil {
		opts = &HandlerOpts{}
	}

	opts.SetSlot()
	return self.subscribe(self.slotDestination(topic), h, opts)
}

func (self *Gilmour) subscribe(topic string, h Handler, opts *HandlerOpts) (*Subscription, error) {
	group := opts.GetGroup()

	if group != "" && self.isExclusiveDuplicate(topic, group) {
		return nil, errors.New(fmt.Sprintf("Duplicate reply handler for %v:%v", topic, group))
	}

	if err := self.backend.Subscribe(topic, opts.GetGroup()); err != nil {
		return nil, err
	} else {
		return self.addSubscriber(topic, h, opts), nil
	}
}

func (self *Gilmour) UnsubscribeSlot(topic string, s *Subscription) {
	self.unsubscribe(self.slotDestination(topic), s)
}

func (self *Gilmour) UnsubscribeReply(topic string, s *Subscription) {
	self.unsubscribe(self.requestDestination(topic), s)
}

func (self *Gilmour) unsubscribe(topic string, s *Subscription) {
	self.removeSubscriber(topic, s)

	if _, ok := self.getSubscribers(topic); !ok {
		err := self.backend.Unsubscribe(topic)
		if err != nil {
			panic(err)
		}
	}
}

func (self *Gilmour) CanReportErrors() bool {
	return self.errorMethod != protocol.BLANK
}

func (self *Gilmour) SetErrorMethod(method string) {
	if method != protocol.QUEUE && method != protocol.PUBLISH && method != protocol.BLANK {
		panic(errors.New(fmt.Sprintf(
			"error method can only be %v, %v or %v",
			protocol.QUEUE, protocol.PUBLISH, protocol.BLANK,
		)))
	}

	self.errorMethod = method
}

func (self *Gilmour) GetErrorMethod() string {
	return self.errorMethod
}

func (self *Gilmour) ReportError(e *protocol.Error) {
	ui.Warn(
		"Reporting Error. Code %v Sender %v Topic %v",
		e.GetCode(), e.GetSender(), e.GetTopic(),
	)

	err := self.backend.ReportError(self.GetErrorMethod(), e)
	if err != nil {
		panic(err)
	}
}

func (self *Gilmour) requestDestination(topic string) string {
	if strings.HasPrefix(topic, "gilmour.") {
		return topic
	} else {
		return fmt.Sprintf("gilmour.request.%v", topic)
	}
}

func (self *Gilmour) slotDestination(topic string) string {
	if strings.HasPrefix(topic, "gilmour.") {
		return topic
	} else {
		return fmt.Sprintf("gilmour.slot.%v", topic)
	}
}

func (self *Gilmour) Request(topic string, msg *Message, opts *RequestOpts) (sender string, err error) {
	if msg == nil {
		msg = NewMessage()
	}

	sender = protocol.MakeSenderId()
	msg.SetSender(sender)

	if opts == nil {
		opts = NewRequestOpts()
	}

	//If a handler is being supplied, subscribe to a response.
	if opts.GetHandler() == nil {
		return sender, errors.New("Cannot use Request without a handler")
	}

	if has, err := self.backend.HasActiveSubscribers(self.requestDestination(topic)); err != nil {
		return sender, err
	} else if !has {
		return sender, errors.New("No active listeners for: " + topic)
	}

	respChannel := self.backend.ResponseTopic(sender)

	//Wait for a responseHandler
	rOpts := NewHandlerOpts().SetOneShot().SetGroup("response")
	self.ReplyTo(respChannel, opts.GetHandler(), rOpts)

	timeout := opts.GetTimeout()
	if timeout > 0 {
		time.AfterFunc(time.Duration(timeout)*time.Second, func() {
			self.sendTimeout(sender, respChannel)
		})
	}

	return sender, self.publish(self.requestDestination(topic), msg)
}

// Same as Request but emulates synchronous behavior. Will not return until you
// have error or data.
func (self *Gilmour) SyncRequest(topic string, msg *Message, opts *RequestOpts) (*Request, error) {
	var req *Request

	var wg sync.WaitGroup
	wg.Add(1)

	if opts == nil {
		opts = NewRequestOpts()
	}

	opts.SetHandler(func(r *Request, _ *Message) {
		defer wg.Done()
		req = r
	})

	_, err := self.Request(topic, msg, opts)
	if err != nil {
		wg.Done()
	}

	wg.Wait()
	return req, err
}

func (self *Gilmour) Signal(topic string, msg *Message) (sender string, err error) {
	if msg == nil {
		msg = NewMessage()
	}

	sender = protocol.MakeSenderId()
	msg.SetSender(sender)
	return sender, self.publish(self.slotDestination(topic), msg)
}

// Internal method to publish a message.
func (self *Gilmour) publish(topic string, msg *Message) error {
	if msg.GetCode() == 0 {
		msg.SetCode(200)
	}

	if msg.GetCode() >= 300 {
		go func() {
			request, err := msg.Marshal()
			if err != nil {
				request = []byte{}
			}

			self.ReportError(
				protocol.MakeError(
					msg.GetCode(),
					topic,
					string(request),
					"",
					msg.GetSender(),
					"",
				),
			)

		}()
	}

	return self.backend.Publish(topic, msg)
}
