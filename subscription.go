package gilmour

type Subscription struct {
	handler     *Handler
	handlerOpts *HandlerOpts
}

func (self *Subscription) GetOpts() *HandlerOpts {
	return self.handlerOpts
}

func (self *Subscription) GetHandler() *Handler {
	return self.handler
}
