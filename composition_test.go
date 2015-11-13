package gilmour

import (
	"sync"
	"testing"
)

func makeCommand() *Command {
	cmd := new(Command)
	cmd.SetTopic(randSeq(5))
	return cmd
}

const (
	composeOne   = "compose-one"
	composeTwo   = "compose-two"
	composeThree = "compose-three"
)

func Setup(e *Gilmour) {
	o := MakeHandlerOpts()

	e.ReplyTo(composeOne, func(r *Request, s *Message) {
		s.Send(map[string]interface{}{"ack-one": "one"})
	}, o)

	e.ReplyTo(composeTwo, func(r *Request, s *Message) {
		s.Send(map[string]interface{}{"ack-two": "two"})
	}, o)

	e.ReplyTo(composeThree, func(r *Request, s *Message) {
		s.Send(map[string]interface{}{"ack-three": "three"})
	}, o)
}

func init() {
	attachSetup(Setup)
}

func makeMessage(data interface{}) *Message {
	m := new(Message)
	m.SetData(data)
	return m
}

func TestCompTransformer(t *testing.T) {
	override := map[string]interface{}{"a": 1, "b": 2}
	m := NewTransformer(override)
	m.Transform(makeMessage("hello"))
}

func TestCompTransformFail(t *testing.T) {
	o := map[string]interface{}{"a": 1, "b": 2}
	m := makeMessage("hello")

	if _, err := NewTransformer(o).Transform(m); err == nil {
		t.Error("Must return error. Cannot merge string and map.")
	}
}

func TestCompTransformMerge(t *testing.T) {
	o := map[string]interface{}{"a": 1, "b": 2}
	data := map[string]interface{}{"x": 1, "y": 2}
	m := makeMessage(data)

	if _, err := NewTransformer(o).Transform(m); err != nil {
		t.Error(err)
	} else if _, ok := data["a"]; !ok {
		t.Error("Must have key a in the merged data")
	}
}

func TestCommand(t *testing.T) {
	cmd := new(Command)
	cmd.SetTopic("hello-world")

	if cmd == nil {
		t.Error("Command cannot be null")
	}

	if cmd.SetTopic("hello-again") == nil {
		t.Error("Cannot set topic twice")
	}
}

func TestCommandTransform(t *testing.T) {
	tr := NewTransformer(map[string]interface{}{"a": 1, "b": 2})
	cmd := new(Command)
	cmd.SetTopic("hello-world")
	cmd.AddTransform(tr)

	if cmd == nil {
		t.Error("Command cannot be null")
	}

	if cmd.AddTransform(tr) == nil {
		t.Error("Cannot set transformation, once already set")
	}
}

func TestComposition(t *testing.T) {
	c := new(Composition)

	cmd := new(Command)
	cmd.SetTopic(composeOne)
	cmd.AddTransform(NewTransformer(map[string]interface{}{"a": 1, "b": 2}))

	c.AddCommand(cmd)

	data := map[string]interface{}{"x": 1, "y": 2}
	m := makeMessage(data)

	var wg sync.WaitGroup
	wg.Add(1)

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		wg.Done()
	})

	engine.Compose(c, m, opts)
	wg.Wait()
}
