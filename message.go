package gilmour

import (
	"encoding/json"
	"sync"

	"gopkg.in/gilmour-libs/gilmour-e-go.v1/protocol"
)

type Message struct {
	data   interface{} `json:"data"`
	code   int         `json:"code"`
	sender string      `json:"sender"`
	sync.Mutex
}

func (self *Message) GetData() interface{} {
	return self.data
}

func (self *Message) Send(data interface{}) {
	self.SetData(data)
}

func (self *Message) SetData(data interface{}) *Message {
	self.Lock()
	defer self.Unlock()

	if self.data != nil {
		panic("Cannot rewrite data for Message.")
	}

	self.data = data
	return self
}

func (self *Message) GetCode() int {
	return self.code
}

func (self *Message) SetCode(code int) *Message {
	self.code = code
	return self
}

func (self *Message) GetSender() string {
	return self.sender
}

func (self *Message) SetSender(sender string) *Message {
	self.sender = sender
	return self
}

func (self *Message) Marshal() ([]byte, error) {
	return json.Marshal(struct {
		Data   interface{} `json:"data"`
		Code   int         `json:"code"`
		Sender string      `json:"sender"`
	}{self.data, self.code, self.sender})
}

func NewMessage() *Message {
	x := &Message{}
	x.SetSender(protocol.MakeSenderId())
	return x
}
