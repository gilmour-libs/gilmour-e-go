package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type GilmourBackend interface {
	Start()
	Stop()

	responseTopic(sender string) string
	subscribe(topic string, opts *HandlerOpts) error
	unsubscribe(topic string) error
	publish(topic string, msg interface{}) error
	reportError(method string, err *protocol.Error) error

	AcquireGroupLock(group string, sender string)
	HealthTopic(ident string) string
	GetIdent() string
	RegisterIdent()
	UnRegisterIdent()
	EnableHealthCheck()
}

type PublishOpts struct {
	Data    interface{}
	Code    int
	Handler *Handler
}
