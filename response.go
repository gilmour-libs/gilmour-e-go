package gilmour

import (
	"errors"
	"sync"
)

type Response struct {
	msgChan chan *Message
	_cap    int
	sent    int
	sync.RWMutex
}

func (r *Response) Next() *Message {
	r.RLock()
	defer r.RUnlock()

	return <-r.msgChan
}

func (r *Response) Cap(m *Message) int {
	r.RLock()
	defer r.RUnlock()

	return r._cap
}

func (r *Response) write(m *Message) error {
	r.Lock()
	defer r.Unlock()

	if r.sent >= r._cap {
		return errors.New("Response buffer overflow")
	}

	r.sent++
	r.msgChan <- m

	if r.sent >= r._cap {
		close(r.msgChan)
	}
	return nil
}

func newResponse(length int) *Response {
	f := make(chan *Message, length)
	return &Response{msgChan: f, _cap: length}
}
