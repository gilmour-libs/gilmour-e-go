package gilmour

import (
	"errors"
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
// Composition.
type Composer interface {
	Execute(*Message, *Gilmour) *Message
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

func (hc *FuncComposition) Execute(m *Message, g *Gilmour) *Message {
	err := compositionMerge(&m.data, &hc.seed)
	if err != nil {
		m := NewMessage()
		m.SetCode(500)
		m.Send(err.Error())
	}

	if m.GetCode() == 0 {
		m.SetCode(200)
	}

	return m
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

func (rc *RequestComposer) Execute(m *Message, g *Gilmour) *Message {
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

	output := <-finally
	if output.GetCode() == 0 {
		output.SetCode(200)
	}

	return output
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
	sync.RWMutex
	mode         string //AndAnd, Pipe, Batch, Parallel
	compositions []Composer
	recordOutput bool
	output       []*Message
}

//Get the output if they were previously recorded, or of a Parallel composition
func (c *Composition) GetOutput() []*Message {
	c.RLock()
	defer c.RUnlock()
	return c.output
}

//Get the list of compositions. Has inbuilt locking.
func (c *Composition) Compositions() []Composer {
	c.RLock()
	defer c.RUnlock()
	return c.compositions
}

//Should the output be recorded? Mostly used in AndAnd and Batch.
func (c *Composition) RecordOutput() *Composition {
	c.Lock()
	defer c.Unlock()

	c.recordOutput = true
	return c
}

// Add a new composer to the List.
func (c *Composition) Add(cp Composer) {
	c.Lock()
	defer c.Unlock()
	c.compositions = append(c.compositions, cp)
}

//Pop the first composer. Useful for Tail recursion.
func (c *Composition) LPop() Composer {
	cmps := c.Compositions()
	c.Lock()
	defer c.Unlock()

	cmd, tail := cmps[0], cmps[1:]
	c.compositions = tail
	return cmd
}

//Save the output message to array of outputs
func (c *Composition) SaveOutput(m *Message) {
	c.Lock()
	defer c.Unlock()

	c.output = append(c.output, m)
}

//Tail recursion over Commands, eventually writing message to requestHandler.
func (c *Composition) do(m *Message, g *Gilmour, cb func(*Message)) {
	//No command left to be executed. Issue the final callback.
	if len(c.Compositions()) == 0 {
		return
	}

	//Update the Tail to only be left with remaining elements.
	cmd := c.LPop()

	//Execute the Composition and pass conrol to the caller
	output := cmd.Execute(copyMessage(m), g)
	if c.recordOutput {
		c.SaveOutput(output)
	}

	cb(output)
}

// Analogus to Linux x && y && z
// Will quit at first failure.
func (c *Composition) andand(m *Message, g *Gilmour, finally chan<- *Message) {
	c.do(m, g, func(msg *Message) {
		if len(c.Compositions()) == 0 {
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
	c.do(m, g, func(msg *Message) {
		if len(c.Compositions()) == 0 {
			finally <- msg
		} else {
			c.batch(m, g, finally)
		}
	})
}

// Analogus to Linux x | y | z
// Will keep going on even if something fails.
func (c *Composition) pipe(m *Message, g *Gilmour, finally chan<- *Message) {
	c.do(m, g, func(msg *Message) {
		if len(c.Compositions()) == 0 {
			finally <- msg
		} else {
			c.pipe(msg, g, finally)
		}
	})
}

// Stops at first successful operation
func (c *Composition) oror(m *Message, g *Gilmour, finally chan<- *Message) {
	c.do(m, g, func(msg *Message) {
		if msg.GetCode() < 400 || len(c.Compositions()) == 0 {
			finally <- msg
		} else {
			c.oror(m, g, finally)
		}
	})
}

func (c *Composition) parallel(m *Message, g *Gilmour, finally chan<- *Message) {
	cmps := c.Compositions()
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
	case parallel:
		c.parallel(m, g, finally)
	case oror:
		c.oror(m, g, finally)
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
func (c *Composition) Execute(m *Message, g *Gilmour) *Message {
	//Wait for the message to show up.
	output := <-c.selectMode(m, g)
	if output.GetCode() == 0 {
		output.SetCode(200)
	}

	return output
}

//New Pipe composition
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

func NewOrOr() *Composition {
	c := new(Composition)
	c.mode = oror
	return c
}

//New Parallel composition
func NewParallel() *Composition {
	c := new(Composition)
	c.mode = parallel
	c.recordOutput = true
	return c
}
