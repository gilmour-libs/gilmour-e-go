package protocol

import (
	"encoding/json"
)

type SentRequest struct {
	Data   interface{} `json:"data"`
	Code   int         `json:"code"`
	Sender string      `json:"sender"`
}

func (self *SentRequest) GetData() interface{} {
	return self.Data
}

func (self *SentRequest) Send(data interface{}) *SentRequest {
	self.Data = data
	return self
}

func (self *SentRequest) GetCode() int {
	return self.Code
}

func (self *SentRequest) SetCode(code int) *SentRequest {
	self.Code = code
	return self
}

func (self *SentRequest) GetSender() string {
	return self.Sender
}

func (self *SentRequest) SetSender(sender string) *SentRequest {
	self.Sender = sender
	return self
}

func (self *SentRequest) Marshal() ([]byte, error) {
	return json.Marshal(self)
}
