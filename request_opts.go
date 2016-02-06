package gilmour

// Structure to handle Reuquest Response.
// Namely, ability to SetTimeout, SetHanler and ConfirmSubscriber().
type RequestOpts struct {
	confirm bool
	timeout int
	handler Handler
}

func (r *RequestOpts) GetTimeout() int {
	return r.timeout
}

func (r *RequestOpts) SetTimeout(t int) *RequestOpts {
	r.timeout = t
	return r
}

func (r *RequestOpts) SetHandler(h Handler) *RequestOpts {
	r.handler = h
	return r
}

func (r *RequestOpts) GetHandler() Handler {
	return r.handler
}

func NewRequestOpts() *RequestOpts {
	return &RequestOpts{}
}
