package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type Gilmer interface {
	Start()
	Stop()
	Subscribe(topic string, h *Handler, opts *HandlerOpts) *Subscription
	UnSubscribe(topic string)
	Publish(topic string, response *protocol.RequestResponse)
	AcquireGroupLock(group string, sender string)
	ResponseTopic(sender string) string
	HealthTopic(ident string) string
	GetIdent() string
	RegisterIdent()
	UnRegisterIdent()
	CanReportErrors() bool
	ReportError(err *protocol.Error)
	EnableHealthCheck()
}
