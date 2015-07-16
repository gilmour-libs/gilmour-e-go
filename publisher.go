package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type Publisher struct {
	data    interface{}
	code    int
	handler Handler
	timeout int
	confirm bool
}

func (self *Publisher) ShouldConfirmSubscriber() bool {
	return self.confirm
}

func (self *Publisher) ConfirmSubscriber() *Publisher {
	self.confirm = true
	return self
}

func (self *Publisher) GetTimeout() int {
	return self.timeout
}

func (self *Publisher) SetTimeout(t int) *Publisher {
	self.timeout = t
	return self
}

func (self *Publisher) SetCode(code int) *Publisher {
	self.code = code
	return self
}

func (self *Publisher) GetCode() int {
	if self.code == 0 {
		self.code = 200
	}

	return self.code
}

func (self *Publisher) SetHandler(h Handler) *Publisher {
	self.handler = h
	return self
}

func (self *Publisher) GetHandler() Handler {
	return self.handler
}

func (self *Publisher) GetData() interface{} {
	return self.data
}

func (self *Publisher) SetData(data interface{}) *Publisher {
	self.data = data
	return self
}

func (self *Publisher) ToSentRequest(sender string) *protocol.SentRequest {
	req := &protocol.SentRequest{}
	return req.SetSender(sender).SetCode(self.GetCode()).Send(self.GetData())
}

func NewPublisher() *Publisher {
	return &Publisher{}
}
