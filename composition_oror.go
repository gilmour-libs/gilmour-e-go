package gilmour

func (g *Gilmour) NewOrOr(cmds ...Executable) *OrOrComposer {
	c := new(OrOrComposer)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

type OrOrComposer struct {
	composition
}

func (c *OrOrComposer) Execute(m *Message) (resp *Response, err error) {
	do := func(do recfunc, m *Message, f *Response) {
		cmd := c.lpop()
		resp, err = performJob(cmd, m)

		// Keep going, If the Pipeline has failed so far and there are still
		// some executables left.
		if len(c.executables()) > 0 && (resp.Code() > 200 || err != nil) {
			do(do, m, f)
			return
		}
	}

	do(do, m, resp)
	return
}
