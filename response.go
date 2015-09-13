package gilmour

import (
	"encoding/json"

	"gopkg.in/gilmour-libs/gilmour-e-go.v0/protocol"
)

type Response struct {
	data   interface{} `json:"data"`
	code   int         `json:"code"`
	sender string      `json:"sender"`
}

func (self *Response) GetData() interface{} {
	return self.data
}

func (self *Response) Send(data interface{}) {
	self.SetData(data)
}

func (self *Response) SetData(data interface{}) *Response {
	if self.data != nil {
		panic("Cannot rewrite data for response.")
	}

	self.data = data
	return self
}

func (self *Response) GetCode() int {
	return self.code
}

func (self *Response) SetCode(code int) *Response {
	self.code = code
	return self
}

func (self *Response) GetSender() string {
	return self.sender
}

func (self *Response) SetSender(sender string) *Response {
	self.sender = sender
	return self
}

func (self *Response) Marshal() ([]byte, error) {
	return json.Marshal(struct {
		Data   interface{} `json:"data"`
		Code   int         `json:"code"`
		Sender string      `json:"sender"`
	}{self.data, self.code, self.sender})
}

func NewResponse() *Response {
	x := &Response{}
	x.SetSender(protocol.MakeSenderId())
	return x
}
