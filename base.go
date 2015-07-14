package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type GilmourBackend interface {
	Start() chan *protocol.Message
	Stop()

	ResponseTopic(sender string) string
	Subscribe(topic string) error
	Unsubscribe(topic string) error
	Publish(topic string, msg interface{}) error
	ReportError(method string, err *protocol.Error) error

	AcquireGroupLock(group, sender string) bool

	RegisterIdent(uuid string) error
	UnregisterIdent(uuid string) error
}

type PublishOpts struct {
	Data    interface{}
	Code    int
	Handler Handler
}
