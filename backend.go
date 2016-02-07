package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-e-go.v4/protocol"
)

type Backend interface {
	Start(chan<- *protocol.Message)
	Stop()

	HasActiveSubscribers(topic string) (bool, error)

	Subscribe(topic, group string) error
	Unsubscribe(topic string) error
	Publish(topic string, msg interface{}) error
	ReportError(method string, err protocol.Error) error

	AcquireGroupLock(group, sender string) bool

	RegisterIdent(uuid string) error
	UnregisterIdent(uuid string) error
}
