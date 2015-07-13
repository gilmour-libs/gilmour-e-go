package protocol

type Messenger interface {
	Marshal() ([]byte, error)
}
