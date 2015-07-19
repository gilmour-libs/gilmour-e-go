package gilmour

import (
	"sync"
)

type Subscriber interface {
	Get(topic string) ([]*Subscription, bool)
	Add(t string, h Handler, o *HandlerOpts) *Subscription
	Delete(topic string, s *Subscription)
	DeleteAll(topic string)
	GetAll() map[string][]*Subscription
}

type Subscription struct {
	handler     Handler
	handlerOpts *HandlerOpts
}

func (self *Subscription) GetOpts() *HandlerOpts {
	if self.handlerOpts == nil {
		self.handlerOpts = MakeHandlerOpts()
	}

	return self.handlerOpts
}

func (self *Subscription) GetHandler() Handler {
	return self.handler
}

type SubscriptionManager struct {
	sync.RWMutex
	hash map[string][]*Subscription
}

func (self *SubscriptionManager) GetAll() map[string][]*Subscription {
	self.RLock()
	defer self.RUnlock()

	return self.hash
}

func (self *SubscriptionManager) Get(topic string) ([]*Subscription, bool) {
	self.RLock()
	defer self.RUnlock()

	list, ok := self.hash[topic]
	return list, ok
}

func (self *SubscriptionManager) Add(t string, h Handler, o *HandlerOpts) *Subscription {
	self.Lock()
	defer self.Unlock()

	if _, ok := self.hash[t]; !ok {
		self.hash[t] = []*Subscription{}
	}

	sub := &Subscription{h, o}

	arr := self.hash[t]
	arr = append(arr, sub)
	self.hash[t] = arr

	return sub

}

func (self *SubscriptionManager) Delete(topic string, s *Subscription) {
	self.Lock()
	defer self.Unlock()

	list, ok := self.hash[topic]
	if !ok {
		return
	}

	new_list := []*Subscription{}

	for _, elem := range list {
		if elem == s {
			//Do nothing
			continue
		}

		new_list = append(new_list, elem)
	}

	self.hash[topic] = new_list
	if len(new_list) == 0 {
		delete(self.hash, topic)
	}

	return
}

func (self *SubscriptionManager) DeleteAll(topic string) {
	self.Lock()
	defer self.Unlock()

	self.hash[topic] = []*Subscription{}
}

func NewSubscriptionManager() *SubscriptionManager {
	x := &SubscriptionManager{}
	x.hash = map[string][]*Subscription{}
	return x
}
