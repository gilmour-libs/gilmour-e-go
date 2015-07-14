package protocol

type Message struct {
	Type  string
	Topic string
	Data  interface{}
	Key   string
}
