package gilmour

import "testing"

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
func TestFuncComposition(t *testing.T) {
	key := "merge"
	data := StrMap{"arg": 1}
	c := NewFuncComposition(StrMap{key: 1})
	_, err := c.Execute(makeMessage(data), engine)
	if err != nil {
		t.Error(err)
	}

	if _, ok := data[key]; !ok {
		t.Error("Must have", key, "in final output")
	}
}

// Transformation should fail while merging string to StrMap
func TestFuncCompositionFail(t *testing.T) {
	data := "arg"
	c := NewFuncComposition(StrMap{"merge": 1})
	_, err := c.Execute(makeMessage(data), engine)
	if err == nil {
		t.Error("Should have raised error")
	}
}

func TestCompositionExecute(t *testing.T) {

	data := StrMap{}
	c := NewRequestComposition(composeOne)
	m, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)

	m.Unmarshal(&data)

	for _, key := range []string{"input", "ack-one"} {
		if _, ok := data[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestCompositionMergeExecute(t *testing.T) {

	data := StrMap{}
	c := NewRequestComposition(composeOne)
	c.SetMessage(StrMap{"merge-one": 1})

	m, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)

	m.Unmarshal(&data)

	for _, key := range []string{"input", "ack-one", "merge-one"} {
		if _, ok := data[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

// Basic composition Test composing single command with single transformation.
// Output must contain the output of the micro service morphed with the
// transformation.
func TestComposePipe(t *testing.T) {
	c := NewPipe()

	c1 := NewRequestComposition(composeOne)
	c1.SetMessage(StrMap{"merge": 1})
	c.Add(c1)

	msg, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)

	expected := StrMap{}
	msg.Unmarshal(&expected)

	if _, ok := expected["merge"]; !ok {
		t.Error("Must have merge in final output")
	}

	if _, ok := expected["ack-one"]; !ok {
		t.Error("Must have ack-one in final output")
	}
}

func TestComposeComplex(t *testing.T) {
	c := NewPipe()

	c1 := NewRequestComposition(composeOne)
	c1.SetMessage(StrMap{"merge-one": 1})
	c.Add(c1)

	c2 := NewRequestComposition(composeTwo)
	c2.SetMessage(StrMap{"merge-two": 1})
	c.Add(c2)

	msg, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)

	expected := StrMap{}
	msg.Unmarshal(&expected)

	for _, key := range []string{"input", "merge-one", "ack-one", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeNested(t *testing.T) {
	c1 := NewPipe()

	c11 := NewRequestComposition(composeOne)
	c11.SetMessage(StrMap{"merge-one": 1})
	c1.Add(c11)

	c2 := NewPipe()

	c2.Add(NewFuncComposition(StrMap{"fake-two": 1}))

	c21 := NewRequestComposition(composeTwo)
	c21.SetMessage(StrMap{"merge-two": 1})
	c2.Add(c21)

	c1.Add(c2)
	msg, _ := c1.Execute(makeMessage(StrMap{"input": 1}), engine)

	expected := StrMap{}
	msg.Unmarshal(&expected)

	for _, key := range []string{"input", "merge-one", "ack-one", "fake-two", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeAndAnd(t *testing.T) {
	c1 := NewAndAnd()

	c11 := NewRequestComposition(composeOne)
	c11.SetMessage(StrMap{"merge-one": 1})
	c1.Add(c11)

	c2 := NewPipe()
	c2.Add(NewFuncComposition(StrMap{"fake-two": 1}))
	c21 := NewRequestComposition(composeTwo)
	c21.SetMessage(StrMap{"merge-two": 1})
	c2.Add(c21)

	c1.Add(c2)
	msg, _ := c1.Execute(makeMessage(StrMap{"input": 1}), engine)

	expected := StrMap{}
	msg.Unmarshal(&expected)

	for _, key := range []string{"merge", "ack-one"} {
		if _, ok := expected[key]; ok {
			t.Error("Must NOT have", key, "in final output")
		}
	}

	for _, key := range []string{"input", "fake-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeAndAndFail(t *testing.T) {
	c := NewAndAnd()

	c.Add(NewRequestComposition(composeOne))
	c.Add(NewRequestComposition(composeBadTwo))
	c.Add(NewRequestComposition(composeTwo))

	msg, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)

	if msg.GetCode() != 500 {
		t.Error("Request should have failed")
	}

	if len(c.compositions) != 1 {
		t.Error("Composition should have been left with more commands.")
	}
}

func TestComposeBatch(t *testing.T) {
	c := NewBatch()

	c1 := NewRequestComposition(composeOne)
	c1.SetMessage(StrMap{"merge": 1})
	c.Add(c1)

	c2 := NewRequestComposition(composeTwo)
	c2.SetMessage(StrMap{"merge-two": 1})
	c.Add(c2)

	msg, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)
	expected := StrMap{}
	msg.Unmarshal(&expected)

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
	c := NewBatch()

	c.Add(NewRequestComposition(composeOne))
	c.Add(NewRequestComposition(composeBadTwo))

	c2 := NewRequestComposition(composeTwo)
	c2.SetMessage(StrMap{"merge-two": 1})
	c.Add(c2)

	msg, _ := c.Execute(makeMessage(StrMap{"input": 1}), engine)
	expected := StrMap{}
	msg.Unmarshal(&expected)

	if msg.GetCode() != 200 {
		t.Error("Request should have passed")
	}

	if len(c.compositions) != 0 {
		t.Error("Composition should not be left with any more commands.")
	}

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
