package gilmour

type HandlerOpts struct {
	group        string
	oneShot      bool
	sendResponse bool
}

func (self *HandlerOpts) GetGroup() string {
	return self.group
}

func (self *HandlerOpts) SetGroup(group string) *HandlerOpts {
	self.group = group
	return self
}

func (self *HandlerOpts) IsSendResponse() bool {
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
	return &HandlerOpts{}
}
