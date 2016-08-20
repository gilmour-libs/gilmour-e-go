package proto

type BackendWriter interface {
	Marshal() ([]byte, error)
}

type BackendPacket struct {
	msg_type string
	topic    string
	data     interface{}
	key      string
}

func (m *BackendPacket) GetTopic() string {
	return m.topic
}

func (m *BackendPacket) GetData() interface{} {
	return m.data
}

func (m *BackendPacket) GetPattern() string {
	return m.key
}

func NewBackendPacket(msg_type, topic, key string, data interface{}) *BackendPacket {
	return &BackendPacket{msg_type, topic, data, key}
}
