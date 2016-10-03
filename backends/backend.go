package backends

import (
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/proto"
)

const (
	ErrorPolicyQueue   = "queue"
	ErrorPolicyPublish = "publish"
	ErrorPolicyIgnore  = ""
)

func ErrorTopic() string {
	return "gilmour.errors"
}

func ErrorQueue() string {
	return "gilmour.errorqueue"
}

type Writer interface {
	Marshal() ([]byte, error)
}

type Backend interface {
	Start(chan<- *proto.Packet)
	Stop()

	HasActiveSubscribers(topic string) (bool, error)

	Subscribe(topic, group string) error
	Unsubscribe(topic string) error
	Publish(topic string, msg interface{}) (bool, error)

	SetErrorPolicy(string) error
	GetErrorPolicy() string
	SupportedErrorPolicies() []string
	ReportError(method string, err *proto.GilmourError) error

	AcquireGroupLock(group, sender string) bool

	RegisterIdent(uuid string) error
	UnregisterIdent(uuid string) error
}
