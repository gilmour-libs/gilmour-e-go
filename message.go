package gilmour

import (
	"encoding/json"
	"sync"
)

type Message struct {
	Data   interface{} `json:"data"`
	Code   int         `json:"code"`
	Sender string      `json:"sender"`
	sync.RWMutex
}

func (m *Message) bytes() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return json.Marshal(m.Data)
}

func (m *Message) SetData(data interface{}) *Message {
	m.Lock()
	defer m.Unlock()

	if m.Data != nil {
		panic("Cannot rewrite data for Message.")
	}

	m.Data = data
	return m
}

func (m *Message) GetData(t interface{}) error {
	m.RLock()
	defer m.RUnlock()

	if byts, err := m.bytes(); err != nil {
		return err
	} else {
		return json.Unmarshal(byts, t)
	}
}

func (m *Message) GetCode() int {
	m.RLock()
	defer m.RUnlock()

	return m.Code
}

func (m *Message) SetCode(code int) *Message {
	m.Lock()
	defer m.Unlock()

	m.Code = code
	return m
}

func (m *Message) GetSender() string {
	m.RLock()
	defer m.RUnlock()

	return m.Sender
}

func (m *Message) setSender(sender string) *Message {
	m.Lock()
	defer m.Unlock()

	m.Sender = sender
	return m
}

func (m *Message) Marshal() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return json.Marshal(m)
}

func parseMessage(data interface{}) (resp *Message, err error) {
	var msg []byte

	switch t := data.(type) {
	case string:
		msg = []byte(t)
	case []byte:
		msg = t
	case json.RawMessage:
		msg = t
	}

	//resp := new(Message)
	err = json.Unmarshal(msg, &resp)
	//if err == nil {
	//	resp = &Message{data: _msg.Data, code: _msg.Code, sender: _msg.Sender}
	//}

	return
}

func NewMessage() *Message {
	x := &Message{}
	x.setSender(makeSenderId())
	return x
}
