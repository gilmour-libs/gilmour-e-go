package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type GilmourBackend interface {
	Start()
	Stop()
	Subscribe(topic string, h *Handler, opts *HandlerOpts) *Subscription
	UnSubscribe(topic string)
	publish(topic string, msg protocol.Messenger)
	AcquireGroupLock(group string, sender string)
	responseTopic(sender string) string
	HealthTopic(ident string) string
	GetIdent() string
	RegisterIdent()
	UnRegisterIdent()
	CanReportErrors() bool
	ReportError(err *protocol.Error)
	EnableHealthCheck()
}

type PublishOpts struct {
	Data    interface{}
	Code    int
	Handler *Handler
}
