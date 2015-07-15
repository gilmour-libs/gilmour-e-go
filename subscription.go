package gilmour

type Subscription struct {
	handler     Handler
	handlerOpts *HandlerOpts
}

func (self *Subscription) GetOpts() *HandlerOpts {
	if self.handlerOpts == nil {
		self.handlerOpts = MakeHandlerOpts()
	}

	return self.handlerOpts
}

func (self *Subscription) GetHandler() Handler {
	return self.handler
}
