package proto

type Packet struct {
	msg_type string
	topic    string
	data     interface{}
	key      string
}

func (m *Packet) GetTopic() string {
	return m.topic
}

func (m *Packet) GetData() interface{} {
	return m.data
}

func (m *Packet) GetPattern() string {
	return m.key
}

func NewPacket(msg_type, topic, key string, data interface{}) *Packet {
	return &Packet{msg_type, topic, data, key}
}
