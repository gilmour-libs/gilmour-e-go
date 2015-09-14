package gilmour

import "sync"

const TIMEOUT = 600

type HandlerOpts struct {
	group   string
	timeout int
	oneShot bool
	isSlot  bool
	sync.Mutex
}

func (self *HandlerOpts) GetTimeout() int {
	//Parallel Goroutines will othrwise run into race condition.
	self.Lock()
	defer self.Unlock()

	if self.timeout != 0 {
		self.timeout = TIMEOUT
	}

	return self.timeout
}

func (self *HandlerOpts) SetTimeout(t int) *HandlerOpts {
	self.timeout = t
	return self
}

func (self *HandlerOpts) GetGroup() string {
	return self.group
}

func (self *HandlerOpts) SetGroup(group string) *HandlerOpts {
	self.group = group
	return self
}

func (self *HandlerOpts) IsOneShot() bool {
	return self.oneShot
}

func (self *HandlerOpts) SetOneShot() *HandlerOpts {
	self.oneShot = true
	return self
}

func (self *HandlerOpts) IsSlot() bool {
	return self.isSlot
}

func (self *HandlerOpts) SetSlot() *HandlerOpts {
	self.isSlot = true
	return self
}

func MakeHandlerOpts() *HandlerOpts {
	return NewHandlerOpts()
}

func NewHandlerOpts() *HandlerOpts {
	return &HandlerOpts{}
}
