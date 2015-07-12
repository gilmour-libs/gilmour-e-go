package gilmour

type HandlerOpts struct {
	group        string
	oneShot      bool
	sendResponse bool
}

func (self *HandlerOpts) GetGroup() string {
	return self.group
}

func (self *HandlerOpts) SetGroup(group string) {
	self.group = group
}

func (self *HandlerOpts) IsSendResponse() bool {
	return self.sendResponse
}

func (self *HandlerOpts) SetSendResponse(send bool) {
	self.sendResponse = send
}

func (self *HandlerOpts) IsOneShot() bool {
	return self.oneShot
}

func (self *HandlerOpts) SetoneShot(oneShot bool) {
	self.oneShot = oneShot
}
