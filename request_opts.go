package gilmour

// Structure to handle Reuquest Response.
// Namely, ability to SetTimeout, SetHanler and ConfirmSubscriber().
type RequestOpts struct {
	confirm bool
	timeout int
}

func (r *RequestOpts) GetTimeout() int {
	return r.timeout
}

func (r *RequestOpts) SetTimeout(t int) *RequestOpts {
	r.timeout = t
	return r
}

func NewRequestOpts() *RequestOpts {
	return &RequestOpts{}
}
