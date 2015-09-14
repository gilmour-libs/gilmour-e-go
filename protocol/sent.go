package protocol

import (
	"encoding/json"
)

type PublishData struct {
	data   interface{} `json:"data"`
	code   int         `json:"code"`
	sender string      `json:"sender"`
}

func (self *PublishData) GetData() interface{} {
	return self.data
}

func (self *PublishData) SetData(data interface{}) *PublishData {
	self.data = data
	return self
}

func (self *PublishData) GetCode() int {
	return self.code
}

func (self *PublishData) SetCode(code int) *PublishData {
	self.code = code
	return self
}

func (self *PublishData) GetSender() string {
	return self.sender
}

func (self *PublishData) SetSender(sender string) *PublishData {
	self.sender = sender
	return self
}

func (self *PublishData) Marshal() ([]byte, error) {
	return json.Marshal(struct {
		Data   interface{} `json:"data"`
		Code   int         `json:"code"`
		Sender string      `json:"sender"`
	}{self.data, self.code, self.sender})
}

func NewPublishData() *PublishData {
	x := &PublishData{}
	x.SetSender(MakeSenderId())
	return x
}
