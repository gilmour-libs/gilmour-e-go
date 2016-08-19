package backends

type Backend interface {
	Start(chan<- MsgReader)
	Stop()

	HasActiveSubscribers(topic string) (bool, error)

	Subscribe(topic, group string) error
	Unsubscribe(topic string) error
	Publish(topic string, msg interface{}) error

	SetErrorPolicy(string) error
	GetErrorPolicy() string
	SupportedErrorPolicies() []string
	ReportError(method string, err MsgWriter) error

	AcquireGroupLock(group, sender string) bool

	RegisterIdent(uuid string) error
	UnregisterIdent(uuid string) error
}
