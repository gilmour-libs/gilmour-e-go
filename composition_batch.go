package gilmour

//New Batch composition
func (g *Gilmour) NewBatch(cmds ...Executable) *BatchComposition {
	c := new(BatchComposition)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

type BatchComposition struct {
	recordableComposition
}

func (c *BatchComposition) Execute(m *Message) (resp *Response, err error) {
	batchResp := c.makeResponse()

	do := func(do recfunc, m *Message, f *Response) {
		cmd := c.lpop()
		resp, err = performJob(cmd, m)

		// Inflate and record the output in a single response.
		if c.isRecorded() {
			r := inflateResponse(resp)
			f.write(r.Next())
		}

		if len(c.executables()) > 0 {
			do(do, m, f)
		}
	}

	do(do, m, batchResp)

	// Automatically return last executable's response or override with
	// recorded output.
	if c.isRecorded() {
		resp = batchResp
	}

	return
}
