package gilmour

import (
	"encoding/json"
	"sync"

	"gopkg.in/gilmour-libs/gilmour-e-go.v5/proto"
)

type pubMsg struct {
	Data   interface{} `json:"data"`
	Code   int         `json:"code"`
	Sender string      `json:"sender"`
}

type Message struct {
	data   interface{}
	code   int
	sender string
	sync.RWMutex
}

func (m *Message) Bytes() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return json.Marshal(m.data)
}

func (m *Message) SetData(data interface{}) *Message {
	m.Lock()
	defer m.Unlock()

	if m.data != nil {
		panic("Cannot rewrite data for Message.")
	}

	m.data = data
	return m
}

func (m *Message) rawData() interface{} {
	m.RLock()
	defer m.RUnlock()

	return m.data
}

func (m *Message) GetData(t interface{}) error {
	m.RLock()
	defer m.RUnlock()

	if byts, err := m.Bytes(); err != nil {
		return err
	} else {
		return json.Unmarshal(byts, t)
	}
}

func (m *Message) GetCode() int {
	m.RLock()
	defer m.RUnlock()

	return m.code
}

func (m *Message) SetCode(code int) *Message {
	m.Lock()
	defer m.Unlock()

	m.code = code
	return m
}

func (m *Message) GetSender() string {
	m.RLock()
	defer m.RUnlock()

	return m.sender
}

func (m *Message) setSender(sender string) *Message {
	m.Lock()
	defer m.Unlock()

	m.sender = sender
	return m
}

func (m *Message) Marshal() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return json.Marshal(pubMsg{m.data, m.code, m.sender})
}

func parseMessage(data interface{}) (*Message, error) {
	resp := &Message{}
	var err error

	if data == nil {
		return resp, err
	}

	var msg []byte

	switch t := data.(type) {
	case string:
		msg = []byte(t)
	case []byte:
		msg = t
	case json.RawMessage:
		msg = t
	}

	_msg := new(pubMsg)
	err = json.Unmarshal(msg, _msg)
	if err == nil {
		resp = &Message{data: _msg.Data, code: _msg.Code, sender: _msg.Sender}
	}

	return resp, err
}

func NewMessage() *Message {
	x := &Message{}
	x.setSender(proto.SenderId())
	return x
}
