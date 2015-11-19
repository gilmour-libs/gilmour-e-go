package gilmour

import (
	"log"
	"sync"
)

const (
	andand   = "andand"
	pipe     = "pipe"
	parallel = "parallel"
	batch    = "batch"
	oror     = "oror"
)

// Common Messenger interface that allows Func Transformer or another
// composition.
type Composer interface {
	Execute(*Message, *Gilmour) <-chan *Message
}

// Standalone Merge transformer.
// Used if you want to transform the output of the previous command before
// seeding it to the next command. Requires to be seeded with an interface
// which will be applied to the previous command's output Message only if the
// type is convertible. In case of a failure it shall raise an Error.
type FuncComposition struct {
	seed interface{}
}

func copyMessage(m *Message) *Message {
	byts, err := m.Marshal()
	if err != nil {
		panic(err)
	}

	msg, err := ParseMessage(byts)
	if err != nil {
		panic(err)
	}

	return msg
}

func outChan() chan *Message {
	return make(chan *Message, 1)
}

func (hc *FuncComposition) Execute(m *Message, g *Gilmour) <-chan *Message {
	err := compositionMerge(&m.data, &hc.seed)
	if err != nil {
		m := NewMessage()
		m.SetCode(500)
		m.Send(err.Error())
	}

	if m.GetCode() == 0 {
		m.SetCode(200)
	}

	finally := outChan()
	finally <- m
	return finally
}

/*
type FunComposer func(interface{}) (*Message, error)

func (hc FunComposer) Execute(m *Message, g *Gilmour) (*Message, error) {
	return hc(m)
}
*/

//Constructor for HashComposer
func NewFuncComposition(s interface{}) *FuncComposition {
	return &FuncComposition{s}
}

// Command represent the each command inside a Pipeline.
// Requires a topic to send the message to, and an optional transformer.
type RequestComposer struct {
	topic   string
	message interface{}
}

func (rc *RequestComposer) With(t interface{}) *RequestComposer {
	if rc.message != nil {
		panic("Cannot change the message after its been set")
	}

	rc.message = t
	return rc
}

func (rc *RequestComposer) Execute(m *Message, g *Gilmour) <-chan *Message {
	if rc.message != nil {
		if err := compositionMerge(&m.data, &rc.message); err != nil {
			log.Println(err)
		}
	}

	finally := outChan()

	opts := NewRequestOpts().SetHandler(func(resp *Request, send *Message) {
		if resp.gData.GetCode() == 0 {
			resp.gData.SetCode(200)
		}
		finally <- resp.gData
	})

	g.Request(rc.topic, m, opts)
	return finally
}

func NewRequestComposition(topic string) *RequestComposer {
	rc := new(RequestComposer)
	rc.topic = topic
	return rc
}

// composition refers to a group of commands and defines the manner in which
// they are to be executed. Common compositions are:
// AndAnd, Batch, Pipe, Parallel. etc. Read documentation for more details.
type composition struct {
	sync.RWMutex
	_compositions []Composer
	output        []*Message
}

//Get the output if they were previously recorded, or of a Parallel composition
func (c *composition) getOutput() []*Message {
	c.RLock()
	defer c.RUnlock()

	return c.output
}

//Get the list of compositions. Has inbuilt locking.
func (c *composition) compositions() []Composer {
	c.RLock()
	defer c.RUnlock()

	return c._compositions
}

// Add a new composer to the List.
func (c *composition) add(cmds ...Composer) {
	c.Lock()
	defer c.Unlock()

	for _, cp := range cmds {
		c._compositions = append(c._compositions, cp)
	}
}

//Pop the first composer. Useful for Tail recursion.
func (c *composition) lpop() Composer {
	cmps := c.compositions()

	c.Lock()
	defer c.Unlock()

	cmd, tail := cmps[0], cmps[1:]
	c._compositions = tail
	return cmd
}

//Save the output message to array of outputs
func (c *composition) saveOutput(m *Message) {
	c.Lock()
	defer c.Unlock()

	c.output = append(c.output, m)
}

type recordableComposition struct {
	composition
	record bool
}

func (c *recordableComposition) isRecorded() bool {
	return c.record
}

//Should the output be recorded? Mostly used in AndAnd and Batch.
func (c *recordableComposition) RecordOutput() {
	c.Lock()
	defer c.Unlock()

	c.record = true
}

func (c *recordableComposition) makeMessage(data interface{}) *Message {
	m := NewMessage()
	m.Send(data)
	m.SetCode(200)
	return m
}

/*
func (c *composition) parallel(m *Message, g *Gilmour, finally chan<- *Message) {
	cmps := c.compositions()
	LEN := len(cmps)
	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup

	outstack := make([]*Message, LEN)

	for ix, cmd := range cmps {
		wg.Add(1)
		go func(cmd Composer, ix int) {
			defer wg.Done()
			output := cmd.Execute(copyMessage(m), g)

			mutex.Lock()
			outstack[ix] = output
			mutex.Unlock()
			log.Println(output, ix)
		}(cmd, ix)
	}

	wg.Wait()
	c.output = outstack
	finally <- outstack[LEN-1]
}

func (c *composition) Do(m *Message, g *Gilmour, o *RequestOpts) {
	finally := make(chan *Message, 1)
	c.Execute(m, g, finally)

	go func() {
		msg := <-finally
		if o == nil {
			return
		}

		fn := o.GetHandler()
		if fn == nil {
			return
		}

		fn(NewRequest("composition", msg), NewMessage())
	}()
}
*/

type recfunc func(recfunc, *Message, *Gilmour, chan<- *Message)

type PipeComposer struct {
	composition
}

func (c *PipeComposer) Execute(m *Message, g *Gilmour) <-chan *Message {
	f := outChan()

	do := func(do recfunc, m *Message, g *Gilmour, f chan<- *Message) {
		cmd := c.lpop()
		finally := cmd.Execute(m, g)

		go func() {
			output := <-finally
			if len(c.compositions()) == 0 || output.GetCode() >= 400 {
				f <- output
			} else {
				do(do, output, g, f)
			}
		}()
	}

	do(do, copyMessage(m), g, f)
	return f
}

type AndAndComposer struct {
	recordableComposition
}

func (c *AndAndComposer) Execute(m *Message, g *Gilmour) <-chan *Message {
	f := outChan()
	outstack := []*Message{}

	do := func(do recfunc, m *Message, g *Gilmour, f chan<- *Message) {
		input := copyMessage(m)
		cmd := c.lpop()
		finally := cmd.Execute(input, g)

		go func() {
			output := <-finally
			if c.isRecorded() {
				outstack = append(outstack, output)
			}
			if len(c.compositions()) == 0 || output.GetCode() >= 400 {
				if c.isRecorded() {
					f <- c.makeMessage(outstack)
				} else {
					f <- output
				}
			} else {
				do(do, m, g, f)
			}
		}()
	}

	do(do, m, g, f)
	return f
}

type OrOrComposer struct {
	composition
}

func (c *OrOrComposer) Execute(m *Message, g *Gilmour) <-chan *Message {
	f := outChan()

	do := func(do recfunc, m *Message, g *Gilmour, f chan<- *Message) {
		input := copyMessage(m)
		cmd := c.lpop()
		finally := cmd.Execute(input, g)

		go func() {
			output := <-finally
			if output.GetCode() < 400 || len(c.compositions()) == 0 {
				f <- output
			} else {
				do(do, m, g, f)
			}
		}()
	}

	do(do, m, g, f)
	return f
}

type BatchComposer struct {
	recordableComposition
}

func (c *BatchComposer) Execute(m *Message, g *Gilmour) <-chan *Message {
	f := outChan()
	outstack := []*Message{}

	do := func(do recfunc, m *Message, g *Gilmour, f chan<- *Message) {
		input := copyMessage(m)
		cmd := c.lpop()
		finally := cmd.Execute(input, g)

		go func() {
			output := <-finally
			if c.isRecorded() {
				outstack = append(outstack, output)
			}
			if len(c.compositions()) == 0 {
				if c.isRecorded() {
					f <- c.makeMessage(outstack)
				} else {
					f <- output
				}
			} else {
				do(do, m, g, f)
			}
		}()
	}

	do(do, m, g, f)
	return f
}

type ParallelComposer struct {
	recordableComposition
}

func (c *ParallelComposer) Execute(m *Message, g *Gilmour) <-chan *Message {
	f := outChan()

	var wg sync.WaitGroup

	mutex := &sync.Mutex{}
	cmps := c.compositions()

	outstack := make([]*Message, len(cmps))

	for ix, cmd := range cmps {
		wg.Add(1)
		go func(cmd Composer, ix int) {
			defer wg.Done()
			output := <-cmd.Execute(copyMessage(m), g)

			mutex.Lock()
			outstack[ix] = output
			mutex.Unlock()

		}(cmd, ix)
	}

	go func() {
		wg.Wait()
		log.Println(c.makeMessage(outstack))
		log.Println(outstack)
		f <- c.makeMessage(outstack)
	}()

	return f
}

//New Pipe composition
func NewPipe(cmds ...Composer) *PipeComposer {
	c := new(PipeComposer)
	c.add(cmds...)
	return c
}

//New AndAnd composition.
func NewAndAnd(cmds ...Composer) *AndAndComposer {
	c := new(AndAndComposer)
	c.add(cmds...)
	return c
}

//New Batch composition
func NewBatch(cmds ...Composer) *BatchComposer {
	c := new(BatchComposer)
	c.add(cmds...)
	return c
}

func NewOrOr(cmds ...Composer) *OrOrComposer {
	c := new(OrOrComposer)
	c.add(cmds...)
	return c
}

//New Parallel composition
func NewParallel(cmds ...Composer) *ParallelComposer {
	c := new(ParallelComposer)
	c.add(cmds...)
	c.RecordOutput()
	return c
}
