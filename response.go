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

func (self *GilmourResponse) isResponseSent() bool {
	return self.responseSent
}

func (self *GilmourResponse) Respond(t interface{}) {
	self.message = t
}

func (self *GilmourResponse) RespondWithCode(t interface{}, code int) {
	self.message = t
	self.code = code
}

func (self *GilmourResponse) Send() (err error) {
	if self.responseSent {
		err = errors.New("Response already sent.")
		return
	}

	self.responseSent = true
	return
}

func NewGilmourResponse(channel string) *GilmourResponse {
	x := GilmourResponse{}
	x.senderchannel = channel
	return &x
}
