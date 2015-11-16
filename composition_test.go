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
	composeOne    = "compose-one"
	composeTwo    = "compose-two"
	composeThree  = "compose-three"
	composeBadTwo = "compose-bad-two"
)

type StrMap map[string]interface{}

func Setup(e *Gilmour) {
	o := MakeHandlerOpts()

	e.ReplyTo(composeBadTwo, func(r *Request, s *Message) {
		panic("bad-two")
	}, o)

	e.ReplyTo(composeOne, func(r *Request, s *Message) {
		req := StrMap{}
		r.Data(&req)
		compositionMerge(&req, &StrMap{"ack-one": "one"})
		s.Send(req)
	}, o)

	e.ReplyTo(composeTwo, func(r *Request, s *Message) {
		req := StrMap{}
		r.Data(&req)
		compositionMerge(&req, &StrMap{"ack-two": "two"})
		s.Send(req)
	}, o)

	e.ReplyTo(composeThree, func(r *Request, s *Message) {
		req := StrMap{}
		r.Data(&req)
		compositionMerge(&req, &StrMap{"ack-three": "three"})
		s.Send(req)
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

// Basic Merger which can be created and exposes a Transform method.
func TestCompMerger(t *testing.T) {
	override := StrMap{"merge": 1}
	m := NewMerger(override)
	m.Transform(makeMessage("hello"))
}

// Transformation should fail while merging string to StrMap
func TestCompTransformFail(t *testing.T) {
	mt := NewMerger(StrMap{"merge": 1})

	m := makeMessage("hello")

	if _, err := mt.Transform(m); err == nil {
		t.Error("Must return error. Cannot merge string and map.")
	}
}

// Transformation must succeed when Seed and Message are of same type.
func TestCompTransformMerge(t *testing.T) {
	mt := NewMerger(StrMap{"merge": 1})

	data := StrMap{"message": 1}
	m := makeMessage(data)

	if _, err := mt.Transform(m); err != nil {
		t.Error(err)
	} else if _, ok := data["merge"]; !ok {
		t.Error("Must have key merge in the merged data")
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
	mt := NewMerger(StrMap{"merge": 1})

	cmd := new(Command)
	cmd.SetTopic("hello-world")
	cmd.AddTransform(mt)

	if cmd == nil {
		t.Error("Command cannot be null")
	}

	if cmd.AddTransform(mt) == nil {
		t.Error("Cannot set transformation, once already set")
	}
}

// Basic composition Test composing single command with single transformation.
// Output must contain the output of the micro service morphed with the
// transformation.
func TestComposePipe(t *testing.T) {
	c := engine.Composition()

	cmd := NewCommand(composeOne)
	cmd.AddTransform(NewMerger(StrMap{"merge": 1}))
	c.AddCommand(cmd)

	var wg sync.WaitGroup
	wg.Add(1)

	expected := StrMap{}

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		defer wg.Done()
		req.Data(&expected)
	})

	c.Execute(makeMessage(StrMap{"input": 1}), opts)
	wg.Wait()

	if _, ok := expected["merge"]; !ok {
		t.Error("Must have merge in final output")
	}

	if _, ok := expected["ack-one"]; !ok {
		t.Error("Must have ack-one in final output")
	}
}

func TestComposeComplex(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	expected := StrMap{}

	c := engine.Composition()

	cmd := NewCommand(composeOne)
	cmd.AddTransform(NewMerger(StrMap{"merge": 1}))
	c.AddCommand(cmd)

	cmd2 := NewCommand(composeTwo)
	cmd2.AddTransform(NewMerger(StrMap{"merge-two": 1}))
	c.AddCommand(cmd2)

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		defer wg.Done()
		req.Data(&expected)
	})

	c.Execute(makeMessage(StrMap{"input": 1}), opts)

	wg.Wait()

	for _, key := range []string{"input", "merge", "ack-one", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeTransform(t *testing.T) {
	c := engine.Composition()

	cmd := NewCommand(composeOne)
	cmd.AddTransform(NewMerger(StrMap{"merge": 1}))
	c.AddCommand(cmd)

	msg, err := c.Transform(makeMessage("ok"))
	if err != nil {
		t.Error("Should not have raised error.")
	}

	expected := StrMap{}
	msg.Unmarshal(&expected)

	for _, key := range []string{"ack-one"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeAndAnd(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	c := engine.AndAnd()

	cmd := NewCommand(composeOne)
	cmd.AddTransform(NewMerger(StrMap{"merge": 1}))
	c.AddCommand(cmd)

	cmd2 := NewCommand(composeTwo)
	cmd2.AddTransform(NewMerger(StrMap{"merge-two": 1}))
	c.AddCommand(cmd2)

	expected := StrMap{}

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		defer wg.Done()
		req.Data(&expected)
	})

	c.Execute(makeMessage(StrMap{"input": 1}), opts)
	wg.Wait()

	for _, key := range []string{"merge", "ack-one"} {
		if _, ok := expected[key]; ok {
			t.Error("Must NOT have", key, "in final output")
		}
	}

	for _, key := range []string{"input", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeAndAndFail(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	c := engine.AndAnd()

	c.AddCommand(NewCommand(composeOne))
	c.AddCommand(NewCommand(composeBadTwo))
	c.AddCommand(NewCommand(composeTwo))

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		defer wg.Done()
		if req.Code() != 500 {
			t.Error("Request should have failed")
		}

		if len(c.cmds) != 1 {
			t.Error("Composition should have been left with more commands.")
		}
	})

	c.Execute(makeMessage(StrMap{"input": 1}), opts)

	wg.Wait()
}

func TestComposeBatch(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	c := engine.Batch()

	cmd := NewCommand(composeOne)
	cmd.AddTransform(NewMerger(StrMap{"merge": 1}))
	c.AddCommand(cmd)

	cmd2 := NewCommand(composeTwo)
	cmd2.AddTransform(NewMerger(StrMap{"merge-two": 1}))
	c.AddCommand(cmd2)

	expected := StrMap{}

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		defer wg.Done()
		req.Data(&expected)
	})

	c.Execute(makeMessage(StrMap{"input": 1}), opts)
	wg.Wait()

	for _, key := range []string{"merge", "ack-one"} {
		if _, ok := expected[key]; ok {
			t.Error("Must NOT have", key, "in final output")
		}
	}

	for _, key := range []string{"input", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeBatchWontFail(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	c := engine.Batch()

	c.AddCommand(NewCommand(composeOne))
	c.AddCommand(NewCommand(composeBadTwo))

	cmd2 := NewCommand(composeTwo)
	cmd2.AddTransform(NewMerger(StrMap{"merge-two": 1}))
	c.AddCommand(cmd2)

	expected := StrMap{}

	opts := NewRequestOpts().SetHandler(func(req *Request, resp *Message) {
		defer wg.Done()
		if req.Code() != 200 {
			t.Error("Request should have passed")
		}

		if len(c.cmds) != 0 {
			t.Error("Composition should not be left with any more commands.")
		}

		req.Data(&expected)
	})

	c.Execute(makeMessage(StrMap{"input": 1}), opts)

	wg.Wait()

	for _, key := range []string{"merge", "ack-one"} {
		if _, ok := expected[key]; ok {
			t.Error("Must NOT have", key, "in final output")
		}
	}

	for _, key := range []string{"input", "ack-two", "merge-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}
