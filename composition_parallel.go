package gilmour

import "sync"

//New Parallel composition
func (g *Gilmour) NewParallel(cmds ...Executable) *ParallelComposer {
	c := new(ParallelComposer)
	c.setEngine(g)
	c.add(cmds...)
	c.RecordOutput()
	return c
}

type ParallelComposer struct {
	recordableComposition
}

func (c *ParallelComposer) Execute(m *Message) (*Response, error) {
	resp := c.makeResponse()
	var wg sync.WaitGroup

	do := func(do recfunc, m *Message, f *Response) {
		wg.Add(1)
		cmd := c.lpop()

		go func(cmd Executable, w *sync.WaitGroup) {
			defer w.Done()
			response, _ := performJob(cmd, m)
			response = inflateResponse(response)
			f.write(response.Next())
		}(cmd, &wg)

		if len(c.executables()) != 0 {
			do(do, m, f)
		}
	}

	do(do, m, resp)
	wg.Wait()
	return resp, nil
}
