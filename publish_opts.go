package gilmour

// Common interface for both Request/Response and Signal/Slot pattern.
type PublishOpts interface {
	ShouldConfirmSubscriber() bool
	GetTimeout() int
	GetHandler() Handler
}

//Structure for Signal Slot Handler.
type SignalOpts struct{}

//Signal Slots do not care about atleast-once delivery.
func (self *SignalOpts) ShouldConfirmSubscriber() bool {
	return false
}

//Signal Slots do not care about delivery, hence no timeout.
func (self *SignalOpts) GetTimeout() int {
	return 0
}

//Signal Slot cannot have a response handler either.
func (self *SignalOpts) GetHandler() Handler {
	return nil
}

// Structure to handle Reuquest Response.
// This one has a few extra methods over and above SignalHadler.
// Namely, ability to SetTimeout, SetHanler and ConfirmSubscriber().
type RequestOpts struct {
	confirm bool
	timeout int
	handler Handler
}

//Override the ShouldConfirmSubscriber to force it to be true.
func (self *RequestOpts) ShouldConfirmSubscriber() bool {
	return true
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
