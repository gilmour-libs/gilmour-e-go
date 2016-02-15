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

/*
Microservices increase the granularity of our services oriented architectures.
Borrowing from unix philosohy, they should do one thing and do it well.
However, for this to be really useful, there should be a facility such as is
provided by unix shells. Unix shells allow the composition of small commands
using the following methods

	- Composition: cmd1 | cmd2 | cmd2
	- AndAnd: cmd1 && cmd2 && cmd3
	- Batch: cmd1; cmd2; cmd3 > out or (cmd1; cmd2; cmd3) > out
	- Parallel: This runs stages in parallel call the callback when all are done.
	The data passed to the callback is an array of :code hashes in no particular
	order. The code is the highest error code from the array.

Also, you should be able to use these methods of composition in any combination
and also in a nexted manner - (cmd1 | cmd2) && cmd3. Gilmour enables you to do
just that.
*/
type Composer interface {
	Execute(*Message) <-chan *Message
	IsStreaming() bool
}

// Standalone Merge transformer.
// Used if you want to transform the output of the previous command before
// seeding it to the next command. Requires to be seeded with an interface
// which will be applied to the previous command's output Message only if the
// type is convertible. In case of a failure it shall raise an Error.
type FuncComposer struct {
	seed interface{}
}

func performJob(cmd Composer, m *Message) *Message {
	jobOutput := cmd.Execute(copyMessage(m))

	var toSend *Message

	if cmd.IsStreaming() {
		// Keep receiving what the channel keeps giving.
		buf := []*Message{}
		code := 200
		for msg := range jobOutput {
			// Code must be set to the highest code in the Output.
			if msg.GetCode() > code {
				code = msg.GetCode()
			}
			buf = append(buf, msg)
		}
		toSend = NewMessage().SetCode(code).SetData(buf)
	} else {
		//Single Message is to be sent.
		toSend = <-jobOutput
	}

	return toSend
}

func copyMessage(m *Message) *Message {
	byts, err := m.Marshal()
	if err != nil {
		panic(err)
	}

	msg, err := parseMessage(byts)
	if err != nil {
		panic(err)
	}

	return msg
}

func outChan() chan *Message {
	return make(chan *Message, 1)
}

func (hc *FuncComposer) IsStreaming() bool {
	return false
}

func (hc *FuncComposer) Execute(m *Message) <-chan *Message {
	err := compositionMerge(&m.Data, &hc.seed)
	if err != nil {
		m := NewMessage()
		m.SetCode(500)
		m.SetData(err.Error())
	}

	if m.GetCode() == 0 {
		m.SetCode(200)
	}

	finally := outChan()
	finally <- m
	close(finally)
	return finally
}

// Command represent the each command inside a Pipeline.
// Requires a topic to send the message to, and an optional transformer.
type RequestComposer struct {
	topic   string
	engine  *Gilmour
	message interface{}
}

func (rc *RequestComposer) IsStreaming() bool {
	return false
}

//Set the Gilmour Engine required for execution
func (rc *RequestComposer) setEngine(g *Gilmour) {
	rc.engine = g
}

func (rc *RequestComposer) With(t interface{}) *RequestComposer {
	if rc.message != nil {
		panic("Cannot change the message after its been set")
	}

	rc.message = t
	return rc
}

func (rc *RequestComposer) Execute(m *Message) <-chan *Message {
	if rc.message != nil {
		if err := compositionMerge(&m.Data, &rc.message); err != nil {
			log.Println(err)
		}
	}

	finally := outChan()

	opts := NewRequestOpts().SetHandler(func(resp *Request, send *Message) {
		if resp.gData.GetCode() == 0 {
			resp.gData.SetCode(200)
		}
		finally <- resp.gData
		close(finally)
	})

	rc.engine.Request(rc.topic, m, opts)
	return finally
}

// composition refers to a group of commands and defines the manner in which
// they are to be executed. Common compositions are:
// AndAnd, Batch, Pipe, Parallel. etc. Read documentation for more details.
type composition struct {
	sync.RWMutex
	engine        *Gilmour
	_compositions []Composer
	output        []*Message
}

func (c *composition) IsStreaming() bool {
	return false
}

//Set the Gilmour Engine required for execution
func (c *composition) setEngine(g *Gilmour) {
	c.engine = g
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

	if len(cmps) == 0 {
		return nil
	}

	cmd := cmps[0]

	if len(cmps) > 1 {
		c._compositions = cmps[1:]
	} else {
		c._compositions = []Composer{}
	}

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

func (c *recordableComposition) IsStreaming() bool {
	return true
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
	m.SetData(data)
	m.SetCode(200)
	return m
}

func (c *recordableComposition) makeChan() (chan *Message, *sync.WaitGroup) {
	var LEN int
	if c.isRecorded() {
		LEN = len(c.compositions())
	} else {
		LEN = 1
	}

	var wg sync.WaitGroup
	wg.Add(len(c.compositions()))

	f := make(chan *Message, LEN)

	//Spin up a goroutine to close the finally channel, when all
	//commands in the composition are done sending their Messages.
	//This is crucial, for the client may be looping over the output channel
	//and any failure to close this chnnel could leave them blocked infinitely.
	if c.isRecorded() {
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(f)
		}(&wg)
	}

	return f, &wg
}

type recfunc func(recfunc, *Message, chan<- *Message)

type PipeComposer struct {
	composition
}

func (c *PipeComposer) Execute(m *Message) <-chan *Message {
	f := outChan()

	do := func(do recfunc, m *Message, f chan<- *Message) {
		cmd := c.lpop()
		toSend := performJob(cmd, m)

		if len(c.compositions()) > 0 && toSend.GetCode() < 400 {
			do(do, toSend, f)
			return
		}

		f <- toSend
		close(f)
	}

	do(do, copyMessage(m), f)
	return f
}

type AndAndComposer struct {
	composition
}

func (c *AndAndComposer) Execute(m *Message) <-chan *Message {
	f := outChan()

	do := func(do recfunc, m *Message, f chan<- *Message) {
		cmd := c.lpop()
		toSend := performJob(cmd, m)

		if len(c.compositions()) > 0 && toSend.GetCode() < 400 {
			do(do, m, f)
			return
		}

		f <- toSend
		close(f)
	}

	do(do, m, f)
	return f
}

type OrOrComposer struct {
	composition
}

func (c *OrOrComposer) Execute(m *Message) <-chan *Message {
	f := outChan()

	do := func(do recfunc, m *Message, f chan<- *Message) {
		cmd := c.lpop()
		toSend := performJob(cmd, m)

		if len(c.compositions()) > 0 && toSend.GetCode() >= 400 {
			do(do, m, f)
			return
		}

		f <- toSend
		close(f)
	}

	do(do, m, f)
	return f
}

type BatchComposer struct {
	recordableComposition
}

func (c *BatchComposer) Execute(m *Message) <-chan *Message {
	f, wg := c.makeChan()

	do := func(do recfunc, m *Message, f chan<- *Message) {
		defer wg.Done()
		cmd := c.lpop()
		terminal := len(c.compositions()) == 0

		toSend := performJob(cmd, m)
		if terminal || c.isRecorded() {
			f <- toSend
		}

		if !terminal {
			do(do, m, f)
		}
	}

	do(do, m, f)
	return f
}

type ParallelComposer struct {
	recordableComposition
}

func (c *ParallelComposer) Execute(m *Message) <-chan *Message {
	f, wg := c.makeChan()

	do := func(do recfunc, m *Message, f chan<- *Message) {
		cmd := c.lpop()
		terminal := len(c.compositions()) == 0

		go func(cmd Composer, terminal bool) {
			defer wg.Done()
			toSend := performJob(cmd, m)
			if terminal || c.isRecorded() {
				f <- toSend
			}
		}(cmd, terminal)

		if !terminal {
			do(do, m, f)
		}
	}

	do(do, m, f)
	return f
}

//Constructor for HashComposer
func (g *Gilmour) NewFuncComposition(s interface{}) *FuncComposer {
	fc := &FuncComposer{s}
	return fc
}

//New Request composition
func (g *Gilmour) NewRequestComposition(topic string) *RequestComposer {
	rc := new(RequestComposer)
	rc.setEngine(g)
	rc.topic = topic
	return rc
}

//New Pipe composition
func (g *Gilmour) NewPipe(cmds ...Composer) *PipeComposer {
	c := new(PipeComposer)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

//New AndAnd composition.
func (g *Gilmour) NewAndAnd(cmds ...Composer) *AndAndComposer {
	c := new(AndAndComposer)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

//New Batch composition
func (g *Gilmour) NewBatch(cmds ...Composer) *BatchComposer {
	c := new(BatchComposer)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

func (g *Gilmour) NewOrOr(cmds ...Composer) *OrOrComposer {
	c := new(OrOrComposer)
	c.setEngine(g)
	c.add(cmds...)
	return c
}

//New Parallel composition
func (g *Gilmour) NewParallel(cmds ...Composer) *ParallelComposer {
	c := new(ParallelComposer)
	c.setEngine(g)
	c.add(cmds...)
	c.RecordOutput()
	return c
}
