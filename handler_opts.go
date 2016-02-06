package gilmour

import "sync"

const TIMEOUT = 600

/*
Handler Options to be passed alongside each Handler at time of topic
subscription. The struct allows chaining and can conveniently be used like

x := NewHandlerOpts().SetTimeout(500).SetGroup("hello-world")
*/
type HandlerOpts struct {
	group   string
	timeout int
	oneShot bool
	_isSlot bool
	sync.RWMutex
}

// Get the execution timeout for the associated handler. If one was not
// explicitly set Gilmour assumes and also sets a default of 600.
func (h *HandlerOpts) GetTimeout() int {
	//Parallel Goroutines will othrwise run into race condition.
	h.Lock()
	defer h.Unlock()

	if h.timeout == 0 {
		h.timeout = TIMEOUT
	}

	return h.timeout
}

// Set the execution timeout for this Handler, otherwise a default of 600
// seconds is assumed.
func (h *HandlerOpts) SetTimeout(t int) *HandlerOpts {
	h.Lock()
	defer h.Unlock()

	h.timeout = t
	return h
}

// If the Handler is to be executed as a part of an exclusive Group.
func (h *HandlerOpts) GetGroup() string {
	h.RLock()
	defer h.RUnlock()

	return h.group
}

// Set this handler to be a part of an exclusive Group. At the most one
// handler is executed per exclusive group.
func (h *HandlerOpts) SetGroup(group string) *HandlerOpts {
	h.Lock()
	defer h.Unlock()

	h.group = group
	return h
}

// Is this a one shot handler? Read Setter for details.
func (h *HandlerOpts) isOneShot() bool {
	h.RLock()
	defer h.RUnlock()

	return h.oneShot
}

// Set this handler to be executed only once. The subscription would be
// unsubscribed on the very first message delivered to the handler,
// either successfully or unsuccessfully.
func (h *HandlerOpts) setOneShot() *HandlerOpts {
	h.Lock()
	defer h.Unlock()

	h.oneShot = true
	return h
}

func (h *HandlerOpts) isSlot() bool {
	h.RLock()
	defer h.RUnlock()

	return h._isSlot
}

func (h *HandlerOpts) setSlot() *HandlerOpts {
	h.Lock()
	defer h.Unlock()

	h._isSlot = true
	return h
}

func NewHandlerOpts() *HandlerOpts {
	return &HandlerOpts{}
}
