package gilmour

import (
	"errors"
	"fmt"
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
	"sync"
)

type Gilmour struct {
	enableHealthCheck bool
	identMutex        sync.RWMutex
	subscriberMutex   sync.RWMutex
	backend           GilmourBackend
	ident             string
	errorMethod       string
	subscribers       map[string][]*Subscription
}

func (self *Gilmour) AddBackend(backend GilmourBackend) {
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

func (self *Gilmour) RegisterIdent() {
	ident := self.GetIdent()
	err := self.backend.RegisterIdent(ident)
	if err != nil {
		panic(err)
	}
}

func (self *Gilmour) UnregisterIdent() {
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
	return self
}

func (self *Gilmour) removeSubscribers(topic string) (err error) {
	self.subscriberMutex.Lock()
	defer self.subscriberMutex.Unlock()

	self.subscribers[topic] = []*Subscription{}
	return
}

func (self *Gilmour) removeSubscriber(topic string, s *Subscription) (length int, err error) {
	self.subscriberMutex.Lock()
	defer self.subscriberMutex.Unlock()

	list, ok := self.subscribers[topic]
	if !ok {
		err = errors.New("Subscribers list is already empty")
		return
	}

	new_list := make([]*Subscription, len(self.subscribers)-1)

	i := 0
	for _, elem := range list {
		if elem == s {
			//Do nothing
			continue
		}

		new_list[i] = elem
		i++
	}

	self.subscribers[topic] = new_list
	length = len(new_list)
	return
}

func (self *Gilmour) addSubscriber(topic string, h Handler, opts *HandlerOpts) *Subscription {
	self.subscriberMutex.Lock()
	defer self.subscriberMutex.Unlock()

	if _, ok := self.subscribers[topic]; !ok {
		self.subscribers[topic] = []*Subscription{}
	}

	sub := &Subscription{h, opts}

	arr := self.subscribers[topic]
	arr = append(arr, sub)

	return sub
}

func (self *Gilmour) Subscribe(topic string, h Handler, opts *HandlerOpts) *Subscription {
	if _, ok := self.subscribers[topic]; !ok {
		self.backend.Subscribe(topic)
	}

	return self.addSubscriber(topic, h, opts)
}

func (self *Gilmour) UnSubscribe(topic string, s *Subscription) {
	var err error

	if err == nil {
		err = self.backend.Unsubscribe(topic)
	}

	if err != nil {
		panic(err)
	}

}

func (self *Gilmour) UnSubscribeAll(topic string) {
	self.removeSubscribers(topic)
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

func (self *Gilmour) ReportError(message *protocol.Error) {
	err := self.backend.ReportError(self.GetErrorMethod(), message)
	if err != nil {
		panic(err)
	}
}

func (self *Gilmour) Publish(topic string, opts *PublishOpts) string {
	//Publish the message

	sender := protocol.MakeSenderId()
	//Always generate a senderId for the message being sent out

	if opts.Handler != nil {
		//If a handler is being supplied, subscribe to a response
		respChannel := self.backend.ResponseTopic(sender)
		//Wait for a responseHandler
		handlerOpts := MakeHandlerOpts().SetOneShot().SetSendResponse(false)
		self.Subscribe(respChannel, opts.Handler, handlerOpts)
	}

	if opts.Code == 0 {
		opts.Code = 200
	}

	message := (&protocol.SentRequest{}).SetSender(sender).SetCode(opts.Code).Send(opts.Data)

	err := self.backend.Publish(topic, message)
	if err != nil {
		panic(err)
	}

	return sender
}

func (self *Gilmour) processMessage(msg *protocol.Message) {
	subs, ok := self.subscribers[msg.Key]
	if !ok || len(subs) == 0 {
		fmt.Println("No subs found!! Key: " + msg.Key)
		return
	}

	for _, s := range subs {
		if s.GetOpts().IsOneShot() {
			self.removeSubscriber(msg.Key, s)
		}

		self.executeSubscriber(s, msg.Topic, msg.Data)
	}
}

func (self *Gilmour) executeSubscriber(s *Subscription, topic string, data interface{}) {
	d, err := protocol.ParseResponse(data)
	if err != nil {
		panic(err)
	}

	opts := s.GetOpts()
	if opts.GetGroup() != protocol.BLANK &&
		!self.backend.AcquireGroupLock(opts.GetGroup(), d.GetSender()) {
		return
	}

	self.handleRequest(s, topic, d)
}

func (self *Gilmour) handleRequest(s *Subscription, topic string, d *protocol.RecvRequest) {
	req := NewGilmourRequest(topic, d)
	res := NewGilmourResponse(self.backend.ResponseTopic(d.GetSender()))
	//GilmourResponder res = new GilmourResponder(backend.responseTopic(d.getSender()));

	//Executing Request
	s.GetHandler()(req, res)
	if s.GetOpts().ShouldSendResponse() {
		var err error

		err = res.Send()
		if err != nil {
			panic(err)
		}

		opts := PublishOpts{res.message, res.code, nil}
		self.Publish(res.senderchannel, &opts)
	}
}

func (self *Gilmour) addConsumer(sink <-chan *protocol.Message) {
	for {
		msg := <-sink
		go self.processMessage(msg)
	}
}

func (self *Gilmour) Start() {
	if self.IsHealthCheckEnabled() {
		self.RegisterIdent()
	}

	sink := self.backend.Start()
	go self.addConsumer(sink)
}

func (self *Gilmour) Stop() {
	if self.IsHealthCheckEnabled() {
		self.UnregisterIdent()
	}

	self.backend.Stop()
}
