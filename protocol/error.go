package protocol

const (
	QUEUE      = "queue"
	PUBLISH    = "publish"
	BLANK      = ""
	ErrorTopic = "gilmour.errors"
)

type Error interface {
	Marshal() ([]byte, error)
}
