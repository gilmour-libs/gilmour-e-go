package protocol

import (
	"encoding/json"
)

type Data struct {
	data   interface{} `json:"data"`
	code   int         `json:"code"`
	sender string      `json:"sender"`
}

func (self *Data) GetData() interface{} {
	return self.data
}

func (self *Data) SetData(data interface{}) *Data {
	self.data = data
	return self
}

func (self *Data) GetCode() int {
	return self.code
}

func (self *Data) SetCode(code int) *Data {
	self.code = code
	return self
}

func (self *Data) GetSender() string {
	return self.sender
}

func (self *Data) SetSender(sender string) *Data {
	self.sender = sender
	return self
}

func (self *Data) Marshal() ([]byte, error) {
	return json.Marshal(self)
}
