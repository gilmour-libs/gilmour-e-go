package gilmour

import (
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
)

type Handler interface {
	process(protocol.Request, protocol.RequestResponse)
}
