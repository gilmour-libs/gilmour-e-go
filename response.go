package gilmour

import (
	"errors"
	"sync"
)

type Response struct {
	msgChan chan *Message
	code    int
	_cap    int
	sent    int
	sync.RWMutex
}

func (r *Response) Next() *Message {
	r.RLock()
	defer r.RUnlock()

	if msg, ok := <-r.msgChan; !ok || msg == nil {
		return nil
	} else {
		return msg
	}
}

func (r *Response) Code() int {
	r.RLock()
	defer r.RUnlock()

	if r.code == 0 {
		r.code = 200
	}

	return r.code
}

func (r *Response) Cap() int {
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

	code := m.GetCode()
	if r.code < code {
		r.code = code
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
