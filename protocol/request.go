package protocol

import (
	"encoding/json"
)

type GilmourRequest struct {
	data   interface{} `json:"data"`
	code   int         `json:"code"`
	sender string      `json:"sender"`
}

func (self *GilmourRequest) GetData() interface{} {
	return self.data
}

func (self *GilmourRequest) SetData(data interface{}) {
	self.data = data
}

func (self *GilmourRequest) GetCode() int {
	return self.code
}

func (self *GilmourRequest) SetCode(code int) {
	self.code = code
}

func (self *GilmourRequest) GetSender() string {
	return self.sender
}

func (self *GilmourRequest) SetSender(sender string) {
	self.sender = sender
}

func (self *GilmourRequest) Render() ([]byte, error) {
	return json.Marshal(self)
}

func MakeRequest(code int, sender string, data string) *GilmourRequest {
	return &GilmourRequest{code: code, data: data, sender: sender}
}
