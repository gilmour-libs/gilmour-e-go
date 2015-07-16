package gilmour

const TIMEOUT = 600

type HandlerOpts struct {
	group        string
	oneShot      bool
	sendResponse bool
	timeout      int
}

func (self *HandlerOpts) GetTimeout() int {
	if self.timeout == 0 {
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

func (self *HandlerOpts) ShouldSendResponse() bool {
	return self.sendResponse
}

func (self *HandlerOpts) SetSendResponse(send bool) *HandlerOpts {
	self.sendResponse = send
	return self
}

func (self *HandlerOpts) IsOneShot() bool {
	return self.oneShot
}

func (self *HandlerOpts) SetOneShot() *HandlerOpts {
	self.oneShot = true
	return self
}

func MakeHandlerOpts() *HandlerOpts {
	opts := HandlerOpts{}
	opts.sendResponse = true
	opts.timeout = TIMEOUT
	return &opts
}
