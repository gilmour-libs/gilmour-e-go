package gilmour

// Structure to handle Reuquest Response.
// Namely, ability to SetTimeout, SetHanler and ConfirmSubscriber().
type RequestOpts struct {
	confirm bool
	timeout int
	handler Handler
}

func (self *RequestOpts) GetTimeout() int {
	return self.timeout
}

func (self *RequestOpts) SetTimeout(t int) *RequestOpts {
	self.timeout = t
	return self
}

func (self *RequestOpts) SetHandler(h Handler) *RequestOpts {
	self.handler = h
	return self
}

func (self *RequestOpts) GetHandler() Handler {
	return self.handler
}

func NewRequestOpts() *RequestOpts {
	return &RequestOpts{}
}
