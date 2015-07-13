package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type Gilmour struct {
	GilmourBackend
}

func (self *Gilmour) Publish(topic string, opts *PublishOpts) string {
	sender := protocol.MakeSenderId()
	if opts.Handler != nil {
		respChannel := self.responseTopic(sender)
		handlerOpts := MakeHandlerOpts().SetOneShot().SetSendResponse(false)
		self.Subscribe(respChannel, opts.Handler, handlerOpts)
	}

	if opts.Code == 0 {
		opts.Code = 200
	}

	message := (&protocol.Data{}).SetSender(sender).SetData(opts.Data).SetCode(opts.Code)

	self.publish(topic, message)
	return sender
}
