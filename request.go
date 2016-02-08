package gilmour

type Request struct {
	topic string
	gData *Message
}

func (r *Request) Sender() string {
	return r.gData.GetSender()
}

func (r *Request) RawData() interface{} {
	return r.gData.data
}

func (r *Request) Data(t interface{}) error {
	return r.gData.Receive(t)
}

func (r *Request) Topic() string {
	return r.topic
}

func (r *Request) Code() int {
	return r.gData.GetCode()
}

func (r *Request) bytes() []byte {
	byt, _ := r.gData.bytes()
	return byt
}

func NewRequest(t string, gd *Message) *Request {
	return &Request{t, gd}
}
