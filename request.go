package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type GilmourRequest struct {
	topic string
	gData *protocol.RecvRequest
}

func (self *GilmourRequest) Sender() string {
	return self.gData.GetSender()
}

func (self *GilmourRequest) Data(t interface{}) {
	self.gData.GetData(t)
}

func (self *GilmourRequest) Topic() string {
	return self.topic
}

func (self *GilmourRequest) Code() int {
	return self.gData.GetCode()
}

func (self *GilmourRequest) StringData() []byte {
	return self.gData.RawData()
}

func NewGilmourRequest(t string, gd *protocol.RecvRequest) *GilmourRequest {
	return &GilmourRequest{t, gd}
}
