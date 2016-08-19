package backends

type MsgWriter interface {
	Marshal() ([]byte, error)
}

type MsgReader interface {
	GetTopic() string
	GetPattern() string
	GetData() interface{}
}

type message struct {
	msg_type string
	topic    string
	data     interface{}
	key      string
}

func (m *message) GetTopic() string {
	return m.topic
}

func (m *message) GetData() interface{} {
	return m.data
}

func (m *message) GetPattern() string {
	return m.key
}
