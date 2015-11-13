package gilmour

import "errors"

// Common Messenger interface that allows Dummy Transformer or another
// Composition.
type Transformer interface {
	Transform(*Message) (*Message, error)
}

// Standalone Method transformer.
// Used if you want to transform the output of the previous command before
// seeding it to the next command. Requires to be seeded with an interface
// which will be applied to the previous command's output Message only if the
// type is convertible. In case of a failure it shall raise an Error.
type FuncTransformer struct {
	seed interface{}
}

func (ft *FuncTransformer) Seed(s interface{}) *FuncTransformer {
	ft.seed = s
	return ft
}

func (ft *FuncTransformer) Transform(m *Message) (*Message, error) {
	err := compositionMerge(&m.data, &ft.seed)
	return m, err
}

func NewTransformer(s interface{}) *FuncTransformer {
	n := new(FuncTransformer)
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

// Composition refers to a group of commands and defines the manner in which
// they are to be executed. Common Compositions are:
// AndAnd, Batch, Pipe, Parallel. etc. Read documentation for more details.
type Composition struct {
	cmds []*Command
}

func (c *Composition) AddCommand(cmd *Command) {
	c.cmds = append(c.cmds, cmd)
}

func (c *Composition) Transform(m *Message) (*Message, error) {
	return new(Message), nil
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
