package gilmour

import (
	"errors"
	"fmt"
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
	"sync"
)

type Gilmour struct {
	subscriberMutex sync.RWMutex
	backend         GilmourBackend
	errorMethod     string
	subscribers     map[string][]*Subscription
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

func (self *Gilmour) addSubscriber(topic string, h *Handler, opts *HandlerOpts) *Subscription {
	self.subscriberMutex.Lock()
	defer self.subscriberMutex.Unlock()

	list, ok := self.subscribers[topic]
	if !ok {
		self.subscribers[topic] = []*Subscription{}
	}

	sub := &Subscription{h, opts}

	arr := self.subscribers[topic]
	arr = append(arr, sub)

	return sub
}

func (self *Gilmour) Subscribe(topic string, h *Handler, opts *HandlerOpts) *Subscription {
	if _, ok := self.subscribers[topic]; !ok {
		self.backend.subscribe(topic, opts)
	}

	return self.addSubscriber(topic, h, opts)
}

func (self *Gilmour) UnSubscribe(topic string, s *Subscription) {
	var err error
	var new_length int

	new_length, err = self.removeSubscriber(topic, s)
	if err == nil {
		err = self.backend.unsubscribe(topic)
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
	err := self.backend.reportError(self.GetErrorMethod(), message)
	if err != nil {
		panic(err)
	}
}

func (self *Gilmour) Publish(topic string, opts *PublishOpts) string {
	sender := protocol.MakeSenderId()
	if opts.Handler != nil {
		respChannel := self.backend.responseTopic(sender)
		handlerOpts := MakeHandlerOpts().SetOneShot().SetSendResponse(false)
		self.Subscribe(respChannel, opts.Handler, handlerOpts)
	}

	if opts.Code == 0 {
		opts.Code = 200
	}

	message := (&protocol.Data{}).SetSender(sender).SetData(opts.Data).SetCode(opts.Code)

	err := self.backend.publish(topic, message)
	if err != nil {
		panic(err)
	}

	return sender
}
