package gilmour

import (
	"sync"
)

//Every subscriptionManager should implement this interface.
type subscriber interface {
	get(topic string) ([]*Subscription, bool)
	add(t string, h handler, o *HandlerOpts) *Subscription
	delete(topic string, s *Subscription)
	deleteAll(topic string)
	getAll() map[string][]*Subscription
}

//Returned for every topic that you subscribe to.
type Subscription struct {
	handler     handler
	handlerOpts *HandlerOpts
}

func (s *Subscription) GetOpts() *HandlerOpts {
	return s.handlerOpts
}

func (s *Subscription) GetHandler() handler {
	return s.handler
}

// Subscriber that maintains a registry of subscriptions for topics.
// It is thread safe.
type subscriptionManager struct {
	sync.RWMutex
	hash map[string][]*Subscription
}

//get all active subscriptions. Returned format is map{topic: [Subscription]}
func (s *subscriptionManager) getAll() map[string][]*Subscription {
	s.RLock()
	defer s.RUnlock()

	return s.hash
}

//get all subscriptions for a particular topic.
func (s *subscriptionManager) get(topic string) ([]*Subscription, bool) {
	s.RLock()
	defer s.RUnlock()

	list, ok := s.hash[topic]
	return list, ok
}

//add a new subscription for a topic. If this topic is being subscribed for
//the first time, initialize an array of subscriptions.
func (s *subscriptionManager) add(t string, h handler, o *HandlerOpts) *Subscription {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.hash[t]; !ok {
		s.hash[t] = []*Subscription{}
	}

	sub := &Subscription{h, o}

	arr := s.hash[t]
	arr = append(arr, sub)
	s.hash[t] = arr

	return sub

}

//Remove a combination of topic & subscription from the manager.
func (s *subscriptionManager) delete(topic string, sub *Subscription) {
	s.Lock()
	defer s.Unlock()

	list, ok := s.hash[topic]
	if !ok {
		return
	}

	new_list := []*Subscription{}

	for _, elem := range list {
		if elem == sub {
			//Do nothing
			continue
		}

		new_list = append(new_list, elem)
	}

	s.hash[topic] = new_list
	if len(new_list) == 0 {
		delete(s.hash, topic)
	}

	return
}

//Delete all subscriptions corresponding to a topic.
func (s *subscriptionManager) deleteAll(topic string) {
	s.Lock()
	defer s.Unlock()

	s.hash[topic] = []*Subscription{}
}

//Constructor
func newSubscriptionManager() *subscriptionManager {
	x := &subscriptionManager{}
	x.hash = map[string][]*Subscription{}
	return x
}
