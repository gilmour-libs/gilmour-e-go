package gilmour

import (
	"errors"
	"log"
)

const (
	andand   = "andand"
	pipe     = "pipe"
	parallel = "parallel"
	batch    = "batch"
)

// Common Messenger interface that allows Func Transformer or another
// Composition.
type Composer interface {
	Execute(*Message, *Gilmour) (*Message, error)
}

// Standalone Merge transformer.
// Used if you want to transform the output of the previous command before
// seeding it to the next command. Requires to be seeded with an interface
// which will be applied to the previous command's output Message only if the
// type is convertible. In case of a failure it shall raise an Error.
type FuncComposition struct {
	seed interface{}
}

func (hc *FuncComposition) Execute(m *Message, g *Gilmour) (*Message, error) {
	err := compositionMerge(&m.data, &hc.seed)
	return m, err
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

func (rc *RequestComposer) SetMessage(t interface{}) (err error) {
	if rc.message != nil {
		err = errors.New("Cannot change the message after its been set")
	} else {
		rc.message = t
	}
	return
}

func (rc *RequestComposer) Execute(m *Message, g *Gilmour) (*Message, error) {
	finally := make(chan *Message, 1)

	if rc.message != nil {
		if err := compositionMerge(&m.data, &rc.message); err != nil {
			log.Println(err)
		}
	}

	opts := NewRequestOpts().SetHandler(func(resp *Request, send *Message) {
		finally <- resp.gData
	})

	g.Request(rc.topic, m, opts)

	return <-finally, nil
}

func NewRequestComposition(topic string) *RequestComposer {
	rc := new(RequestComposer)
	rc.topic = topic
	return rc
}

// Composition refers to a group of commands and defines the manner in which
// they are to be executed. Common Compositions are:
// AndAnd, Batch, Pipe, Parallel. etc. Read documentation for more details.
type Composition struct {
	mode         string //AndAnd, Pipe, Batch, Parallel
	compositions []Composer
}

func (c *Composition) Add(cp Composer) {
	c.compositions = append(c.compositions, cp)
}

//Tail recursion over Commands, eventually writing message to requestHandler.
func (c *Composition) do(m *Message, g *Gilmour, cb func(*Message, error)) {
	//No command left to be executed. Issue the final callback.
	if len(c.compositions) == 0 {
		return
	}

	cmd, tail := c.compositions[0], c.compositions[1:]

	byts, err := m.Marshal()
	if err != nil {
		panic(err)
	}

	msg, err := ParseMessage(byts)
	if err != nil {
		panic(err)
	}

	//Update the Tail to only be left with remaining elements.
	c.compositions = tail

	//Execute the Composition and pass conrol to the caller
	cb(cmd.Execute(msg, g))
}

func copyReqMsg(r *Request) (*Message, error) {
	if byts, err := r.gData.Marshal(); err != nil {
		return nil, err
	} else {
		return ParseMessage(byts)
	}
}

// Analogus to Linux x && y && z
// Will quit at first failure.
func (c *Composition) andand(m *Message, g *Gilmour, finally chan<- *Message) {
	c.do(m, g, func(msg *Message, e error) {
		if len(c.compositions) == 0 {
			finally <- msg
		} else if msg.GetCode() >= 400 {
			//AndAnd should fail on first failure.
			finally <- msg
		} else {
			c.andand(m, g, finally)
		}
	})
}

// Analogus to Linux x; y; z
// Will not stop at any failure.
func (c *Composition) batch(m *Message, g *Gilmour, finally chan<- *Message) {
	c.do(m, g, func(msg *Message, e error) {
		if len(c.compositions) == 0 {
			finally <- msg
		} else {
			c.batch(m, g, finally)
		}
	})
}

// Analogus to Linux x | y | z
// Will keep going on even if something fails.
func (c *Composition) pipe(m *Message, g *Gilmour, finally chan<- *Message) {
	c.do(m, g, func(msg *Message, e error) {
		if len(c.compositions) == 0 {
			finally <- msg
		} else {
			c.pipe(msg, g, finally)
		}
	})
}

func (c *Composition) selectMode(m *Message, g *Gilmour) <-chan *Message {
	//Buffered channel to wait for the message to show up.
	finally := make(chan *Message, 1)

	switch c.mode {
	case andand:
		c.andand(m, g, finally)
	case batch:
		c.batch(m, g, finally)
	case pipe:
		c.pipe(m, g, finally)
	default:
		panic("Unsupported Composition mode")
	}

	return finally
}

func (c *Composition) Do(m *Message, g *Gilmour, o *RequestOpts) {
	finally := c.selectMode(m, g)

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

// Transformer compliant method to be able to pass compositions as valid
// transformations to commands.
func (c *Composition) Execute(m *Message, g *Gilmour) (*Message, error) {
	finally := c.selectMode(m, g)

	//Wait for the message to show up.
	return <-finally, nil
}

type CompositionOpts struct {
	handler      Handler
	recordOutput bool
}

/*
//Override the ShouldConfirmSubscriber to force it to be true.
func (self *CompositionOpts) ShouldRecordOutput() bool {
	return self.recordOutput
}

func (self *CompositionOpts) SetHandler(h Handler) {
	self.handler = h
}

func (self *CompositionOpts) GetHandler() Handler {
	return self.handler
}

func NewPipeOpts() *CompositionOpts {
	return &CompositionOpts{}
}
*/

//New Parallel composition
func NewPipe() *Composition {
	c := new(Composition)
	c.mode = pipe
	return c
}

//New AndAnd Composition.
func NewAndAnd() *Composition {
	c := new(Composition)
	c.mode = andand
	return c
}

//New Batch composition
func NewBatch() *Composition {
	c := new(Composition)
	c.mode = batch
	return c
}

//New Parallel composition
func NewParallel() *Composition {
	c := new(Composition)
	c.mode = parallel
	return c
}
