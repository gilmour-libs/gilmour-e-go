package gilmour

type Request struct {
	topic string
	gData *Message
}

func (self *Request) Sender() string {
	return self.gData.GetSender()
}

func (self *Request) RawData() interface{} {
	return self.gData.data
}

func (self *Request) Data(t interface{}) {
	self.gData.Unmarshal(t)
}

func (self *Request) Topic() string {
	return self.topic
}

func (self *Request) Code() int {
	return self.gData.GetCode()
}

func (self *Request) StringData() []byte {
	byt, _ := self.gData.StringData()
	return byt
}

func NewRequest(t string, gd *Message) *Request {
	return &Request{t, gd}
}
