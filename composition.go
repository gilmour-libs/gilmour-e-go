package gilmour

import (
	"errors"
	"log"
)

const (
	AndAnd   = "andand"
	Pipe     = "pipe"
	Parallel = "parallel"
	Batch    = "batch"
)

// Common Messenger interface that allows Dummy Transformer or another
// Composition.
type Transformer interface {
	Transform(*Message) (*Message, error)
}

// Standalone Merge transformer.
// Used if you want to transform the output of the previous command before
// seeding it to the next command. Requires to be seeded with an interface
// which will be applied to the previous command's output Message only if the
// type is convertible. In case of a failure it shall raise an Error.
type MergeTransform struct {
	seed interface{}
}

func (ft *MergeTransform) Seed(s interface{}) *MergeTransform {
	ft.seed = s
	return ft
}

func (ft *MergeTransform) Transform(m *Message) (*Message, error) {
	err := compositionMerge(&m.data, &ft.seed)
	return m, err
}

func NewMerger(s interface{}) *MergeTransform {
	n := new(MergeTransform)
	n.Seed(s)
	return n
}

// Command represent the each command inside a Pipeline.
// Requires a topic to send the message to, and an optional transformer.
type Command struct {
	topic       string
	transformer Transformer
}

func (c *Command) SetTopic(t string) (err error) {
	if c.topic != "" {
		err = errors.New("Cannot change the topic after its been set")
	} else {
		c.topic = t
	}

	return
}

func (c *Command) AddTransform(t Transformer) (err error) {
	if c.transformer != nil {
		err = errors.New("Cannot change the transformation after its been set")
	} else {
		c.transformer = t
	}
	return
}

func NewCommand(topic string) *Command {
	x := new(Command)
	x.SetTopic(topic)
	return x
}

// Composition refers to a group of commands and defines the manner in which
// they are to be executed. Common Compositions are:
// AndAnd, Batch, Pipe, Parallel. etc. Read documentation for more details.
type Composition struct {
	mode   string //AndAnd, Pipe, Batch, Parallel
	engine *Gilmour
	cmds   []*Command
}

func (c *Composition) AddCommand(cmd *Command) {
	c.cmds = append(c.cmds, cmd)
}

//Tail recursion over Commands, eventually writing message to requestHandler.
func (c *Composition) internal_do(m *Message, do func(*Request, *Message)) {
	//No command left to be executed. Issue the final callback.
	if len(c.cmds) == 0 {
		return
	}

	cmd, tail := c.cmds[0], c.cmds[1:]

	byts, err := m.Marshal()
	if err != nil {
		panic(err)
	}

	msg, err := ParseMessage(byts)
	if err != nil {
		panic(err)
	}

	if cmd.transformer != nil {
		if _, err := cmd.transformer.Transform(msg); err != nil {
			log.Println(err)
		}
	}

	//Update the Tail to only be left with remaining elements.
	c.cmds = tail
	opts := NewRequestOpts().SetHandler(do)
	c.engine.Request(cmd.topic, msg, opts)
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
func (c *Composition) andand(m *Message, finally chan<- *Message) {
	c.internal_do(m, func(r *Request, s *Message) {
		msg, err := copyReqMsg(r)
		if err != nil {
			panic(err)
		}

		if len(c.cmds) == 0 {
			finally <- msg
		} else if msg.GetCode() >= 400 {
			//AndAnd should fail on first failure.
			finally <- msg
		} else {
			c.andand(m, finally)
		}
	})
}

// Analogus to Linux x; y; z
// Will not stop at any failure.
func (c *Composition) batch(m *Message, finally chan<- *Message) {
	c.internal_do(m, func(r *Request, s *Message) {
		msg, err := copyReqMsg(r)
		if err != nil {
			panic(err)
		}

		if len(c.cmds) == 0 {
			finally <- msg
		} else {
			c.batch(m, finally)
		}
	})
}

// Analogus to Linux x | y | z
// Will keep going on even if something fails.
func (c *Composition) pipe(m *Message, finally chan<- *Message) {
	c.internal_do(m, func(r *Request, s *Message) {
		msg, err := copyReqMsg(r)
		if err != nil {
			panic(err)
		}

		if len(c.cmds) == 0 {
			finally <- msg
		} else {
			c.pipe(msg, finally)
		}
	})
}

func (c *Composition) selectMode(m *Message) <-chan *Message {
	//Buffered channel to wait for the message to show up.
	finally := make(chan *Message, 1)

	switch c.mode {
	case AndAnd:
		c.andand(m, finally)
	case Batch:
		c.batch(m, finally)
	case Pipe:
		c.pipe(m, finally)
	default:
		panic("Unsupported Composition mode")
	}

	return finally
}

func (c *Composition) Execute(m *Message, o *RequestOpts) {
	finally := c.selectMode(m)

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
func (c *Composition) Transform(m *Message) (*Message, error) {
	finally := c.selectMode(m)

	//Wait for the message to show up.
	return <-finally, nil
}

type CompositionOpts struct {
	handler      Handler
	recordOutput bool
}

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

func NewCompositionOpts() *CompositionOpts {
	return &CompositionOpts{}
}
