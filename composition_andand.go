package gilmour

//New AndAnd composition.
func (g *Gilmour) NewAndAnd(cmds ...Executable) *AndAndComposer {
	c := new(AndAndComposer)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

type AndAndComposer struct {
	composition
}

func (c *AndAndComposer) Execute(m *Message) (resp *Response, err error) {
	do := func(do recfunc, m *Message, r *Response) {
		cmd := c.lpop()
		resp, err = performJob(cmd, m)

		// Keep going if nothing has failed so far.
		if len(c.executables()) > 0 && err == nil && resp.Code() == 200 {
			do(do, m, r)
			return
		}
	}

	do(do, m, resp)
	return
}
