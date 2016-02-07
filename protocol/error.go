package protocol

const (
	ErrorPolicyQueue   = "queue"
	ErrorPolicyPublish = "publish"
	ErrorPolicyIgnore  = ""
	ErrorTopic         = "gilmour.errors"
)

type Error interface {
	Marshal() ([]byte, error)
}
