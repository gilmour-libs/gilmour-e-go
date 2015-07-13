package gilmour

import (
	"fmt"
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

const QUEUE = "queue"
const PUBLISH = "publish"
const BLANK = ""

const defaultErrorQueue = "gilmour.errorqueue"
const defaultErrorTopic = "gilmour.errors"
const defaultHealthTopic = "gilmour.health"
const defaultResponseTopic = "gilmour.response"
const defaultIdentKey = "gilmour.known_host.health"
const defaultErrorBuffer = 9999

type Redis struct {
	errorMethod string
}

func (self *Redis) getErrorTopic() string {
	return defaultErrorTopic
}

func (self *Redis) getErrorQueue() string {
	return defaultErrorQueue
}

func (self *Redis) getErrorMethod() string {
	return self.errorMethod
}

func (self *Redis) setErrorMethod(method string) {
	self.errorMethod = method
}

func (self *Redis) CanReportErrors() bool {
	return self.errorMethod != BLANK
}

func (self *Redis) ReportError(message *protocol.Error) {
	switch self.errorMethod {
	case PUBLISH:
		self.publish(defaultErrorTopic, message)
	case QUEUE:
		self.redis.lpush(this.errorQueue, message)
		self.redis.ltrim(this.errorQueue, 0, defaultErrorBuffer)
	default:
		fmt.Println("Do not do anything")
	}
}

func (self *Redis) Subscribe(topic string, h *Handler, opts *HandlerOpts) *Subscription {
}

func (self *Redis) responseTopic(sender string) string {
	return defaultResponseTopic + "." + sender
}

func (self *Redis) publish(topic string, message protocol.Messenger) {
	fmt.Println(message.Marshal())
}
