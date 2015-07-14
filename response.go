package gilmour

import (
	"errors"
)

type GilmourResponse struct {
	senderchannel string
	message       interface{}
	code          int
	responseSent  bool
}

func (self Response) isResponseSent() bool {
	return self.responseSent
}

func (self Response) Respond(t interface{}) {
	self.message = t
}

func (self Response) RespondWithCode(t interface{}, code int) {
	self.message = t
	self.code = code
}

func (self Response) Send() (err error) {
	if self.responseSent {
		err = errors.New("Response already sent.")
		return
	}

	self.responseSent = true
	return
}

func NewGilmourResponse(channel string) Response {
	x := GilmourResponse{}
	x.senderchannel = channel
	return &x
}
