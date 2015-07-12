package protocol

import (
	"encoding/json"
)

type Request struct {
	data   interface{} `json:"data"`
	code   int         `json:"code"`
	sender string      `json:"sender"`
}

func (self *Request) GetData() interface{} {
	return self.data
}

func (self *Request) SetData(data interface{}) {
	self.data = data
}

func (self *Request) GetCode() int {
	return self.code
}

func (self *Request) SetCode(code int) {
	self.code = code
}

func (self *Request) GetSender() string {
	return self.sender
}

func (self *Request) SetSender(sender string) {
	self.sender = sender
}

func (self *Request) Render() ([]byte, error) {
	return json.Marshal(self)
}

func MakeRequest(code int, sender string, data string) *Request {
	return &Request{code: code, data: data, sender: sender}
}
