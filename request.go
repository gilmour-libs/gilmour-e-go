package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type Request struct {
	topic string
	gData *protocol.RecvRequest
}

func (self *Request) Sender() string {
	return self.gData.GetSender()
}

func (self *Request) Data(t interface{}) {
	self.gData.GetData(t)
}

func (self *Request) Topic() string {
	return self.topic
}

func (self *Request) Code() int {
	return self.gData.GetCode()
}

func (self *Request) StringData() []byte {
	return self.gData.RawData()
}

func NewRequest(t string, gd *protocol.RecvRequest) *Request {
	return &Request{t, gd}
}
