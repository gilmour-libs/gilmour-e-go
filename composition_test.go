package gilmour

import "testing"

const (
	topicOne    = "compose-one"
	topicTwo    = "compose-two"
	topicThree  = "compose-three"
	topicBadTwo = "compose-bad-two"
)

type StrMap map[string]interface{}

func setup(e *Gilmour) {
	o := NewHandlerOpts()

	e.ReplyTo(topicBadTwo, func(r *Request, s *Message) {
		panic("bad-two")
	}, o)

	e.ReplyTo(topicOne, func(r *Request, s *Message) {
		req := StrMap{}
		r.Data(&req)
		compositionMerge(&req, &StrMap{"ack-one": "one"})
		s.SetData(req)
	}, o)

	e.ReplyTo(topicTwo, func(r *Request, s *Message) {
		req := StrMap{}
		r.Data(&req)
		compositionMerge(&req, &StrMap{"ack-two": "two"})
		s.SetData(req)
	}, o)

	e.ReplyTo(topicThree, func(r *Request, s *Message) {
		req := StrMap{}
		r.Data(&req)
		compositionMerge(&req, &StrMap{"ack-three": "three"})
		s.SetData(req)
	}, o)
}

func init() {
	attachSetup(setup)
}

func makeMessage(data interface{}) *Message {
	m := new(Message)
	m.SetData(data)
	return m
}

// Basic Merger which can be created and exposes a Transform method.
func TestFuncComposition(t *testing.T) {
	key := "merge"

	c := engine.NewLambda(func(m *Message) (*Message, error) {
		var d StrMap
		m.GetData(&d)
		d[key] = 1
		return (&Message{}).SetCode(200).SetData(d), nil
	})

	resp, err := c.Execute(makeMessage(StrMap{"arg": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	if msg.GetCode() != 200 {
		t.Error("Should not have raised Error")
	}

	var data StrMap
	msg.GetData(&data)

	if _, ok := data[key]; !ok {
		t.Error("Must have", key, "in final output")
	}
}

func TestCompositionExecute(t *testing.T) {

	data := StrMap{}
	c := engine.NewRequest(topicOne)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	msg.GetData(&data)

	for _, key := range []string{"input", "ack-one"} {
		if _, ok := data[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestCompositionMergeExecute(t *testing.T) {
	data := StrMap{}
	c := engine.NewRequest(topicOne).With(StrMap{"merge-one": 1})

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	m := resp.Next()
	m.GetData(&data)

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
	c := engine.NewPipe(engine.NewRequest(topicOne).With(StrMap{"merge": 1}))

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

	if _, ok := expected["merge"]; !ok {
		t.Error("Must have merge in final output")
	}

	if _, ok := expected["ack-one"]; !ok {
		t.Error("Must have ack-one in final output")
	}
}

func TestComposeComplex(t *testing.T) {
	c := engine.NewPipe(
		engine.NewRequest(topicOne).With(StrMap{"merge-one": 1}),
		engine.NewRequest(topicTwo).With(StrMap{"merge-two": 1}),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

	for _, key := range []string{"input", "merge-one", "ack-one", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeNested(t *testing.T) {
	c11 := engine.NewRequest(topicOne).With(StrMap{"merge-one": 1})

	c2 := engine.NewPipe(
		engine.NewLambda(func(m *Message) (*Message, error) {
			var d StrMap
			m.GetData(&d)
			d["fake-two"] = 1
			return (&Message{}).SetData(d), nil
		}),
		engine.NewRequest(topicTwo).With(StrMap{"merge-two": 1}),
	)

	c1 := engine.NewPipe(c11, c2)

	resp, err := c1.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

	for _, key := range []string{"input", "merge-one", "ack-one", "fake-two", "merge-two", "ack-two"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}
}

func TestComposeAndAnd(t *testing.T) {
	c2 := engine.NewPipe(
		engine.NewLambda(func(m *Message) (*Message, error) {
			var d StrMap
			m.GetData(&d)
			d["fake-two"] = 1
			return (&Message{}).SetData(d), nil
		}),
		engine.NewRequest(topicTwo).With(StrMap{"merge-two": 1}),
	)

	c1 := engine.NewAndAnd(
		engine.NewRequest(topicOne).With(StrMap{"merge-one": 1}),
		c2,
	)

	resp, err := c1.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

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
	c := engine.NewAndAnd(
		engine.NewRequest(topicOne),
		engine.NewRequest(topicBadTwo),
		engine.NewRequest(topicTwo),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	if msg.GetCode() != 500 {
		t.Error("Request should have failed")
	}

	if len(c.executables()) != 1 {
		t.Error("Composition should have been left with more commands.")
	}
}

func TestComposeBatchPass(t *testing.T) {
	c := engine.NewBatch(
		engine.NewRequest(topicOne).With(StrMap{"merge": 1}),
		engine.NewRequest(topicTwo).With(StrMap{"merge-two": 1}),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

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
	c := engine.NewBatch(
		engine.NewRequest(topicOne),
		engine.NewRequest(topicBadTwo),
		engine.NewRequest(topicTwo).With(StrMap{"merge-two": 1}),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

	if msg.GetCode() != 200 {
		t.Error("Request should have passed")
	}

	if len(c.executables()) != 0 {
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

func TestBatch(t *testing.T) {
	c := engine.NewBatch(
		engine.NewRequest(topicOne),
		engine.NewRequest(topicTwo),
		engine.NewRequest(topicThree),
	)

	f, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should nor have raised Error")
		return
	}

	if f.Cap() != 1 {
		t.Error("Must only capture 1 output.")
	}

	m := f.Next()
	expected := StrMap{}
	if err := m.GetData(&expected); err != nil {
		t.Error("Must be valid message output")
	} else if _, ok := expected["input"]; !ok {
		t.Error("Must have input in final output")
	}
}

func TestBatchRecordOutput(t *testing.T) {
	c := engine.NewBatch(
		engine.NewRequest(topicOne),
		engine.NewRequest(topicTwo),
		engine.NewRequest(topicThree),
	)
	c.RecordOutput()

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	out := []*Message{}
	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out = append(out, msg)
	}

	if len(out) != 3 {
		t.Error("Must have captured 3 outputs.")
	}

	for _, m := range out {
		expected := StrMap{}
		if err := m.GetData(&expected); err != nil {
			t.Error("Must be valid message output")
		} else if _, ok := expected["input"]; !ok {
			t.Error("Must have input in final output")
		}
	}
}

func TestBatchBadRecord(t *testing.T) {
	c := engine.NewBatch(
		engine.NewRequest(topicOne),
		engine.NewRequest(topicBadTwo),
		engine.NewRequest(topicThree),
	)
	c.RecordOutput()

	f, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	if f.Cap() != 3 {
		t.Error("Must have captured 3 outputs.")
	}

	i := 0
	for msg := f.Next(); msg != nil; msg = f.Next() {
		if i != 1 {
			continue
		}
		if msg.GetCode() != 500 {
			t.Error("Should have failed with 500")
		}
	}
}

func TestOrOr(t *testing.T) {
	c := engine.NewOrOr(
		engine.NewRequest(topicBadTwo),
		engine.NewRequest(topicOne),
		engine.NewRequest(topicThree),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	msg := resp.Next()
	expected := StrMap{}
	msg.GetData(&expected)

	for _, key := range []string{"ack-three"} {
		if _, ok := expected[key]; ok {
			t.Error("Must NOT have", key, "in final output")
		}
	}

	for _, key := range []string{"input", "ack-one"} {
		if _, ok := expected[key]; !ok {
			t.Error("Must have", key, "in final output")
		}
	}

}

func TestParallel(t *testing.T) {
	c := engine.NewParallel(
		engine.NewRequest(topicOne),
		engine.NewRequest(topicTwo),
		engine.NewRequest(topicThree),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	if resp.Cap() != 3 {
		t.Error("Must have captured 3 outputs.")
	}

	out := []*Message{}
	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out = append(out, msg)
	}

	if len(out) != 3 {
		t.Error("Must have captured 3 outputs.")
	}

	for _, msg := range out {
		data := StrMap{}
		msg.GetData(&data)
		for _, key := range []string{"input"} {
			if _, ok := data[key]; !ok {
				t.Error("Must have", key, "in final output")
			}
		}
	}
}

func TestOrOrParallel(t *testing.T) {
	c := engine.NewOrOr(
		engine.NewRequest(topicBadTwo),
		engine.NewParallel(
			engine.NewRequest(topicOne),
			engine.NewRequest(topicTwo),
			engine.NewRequest(topicThree),
		),
		engine.NewRequest(topicOne),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	if resp.Cap() != 3 {
		t.Error("Must return 3 outputs.")
		return
	}

	out := []*Message{}
	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out = append(out, msg)
	}

	if len(out) != 3 {
		t.Error("Parallel should have captured 3 outputs.")
		return
	}

	expected := StrMap{}
	out[0].GetData(&expected)

	if _, ok := expected["input"]; !ok {
		t.Error("Must have key: input in final output")
		return
	}
}

func TestParallelParallel(t *testing.T) {
	c := engine.NewParallel(
		engine.NewParallel(
			engine.NewRequest(topicOne).With(StrMap{"1a": 1}),
			engine.NewRequest(topicTwo).With(StrMap{"2a": 1}),
			engine.NewRequest(topicThree).With(StrMap{"3a": 1}),
		),
		engine.NewParallel(
			engine.NewRequest(topicOne).With(StrMap{"1b": 1}),
			engine.NewRequest(topicTwo).With(StrMap{"2b": 1}),
			engine.NewRequest(topicThree).With(StrMap{"3b": 1}),
		),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	if resp.Cap() != 2 {
		t.Error("Must capture 2 outputs.")
		return
	}

	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out := []StrMap{}
		msg.GetData(&out)

		if len(out) != 3 {
			t.Error("Must have captured 3 outputs.")
		}

		for _, x := range out {
			if _, ok := x["input"]; !ok {
				t.Error("Must have key: input in final output")
			}
		}
	}
}

func TestAndAndParallel(t *testing.T) {
	c := engine.NewAndAnd(
		engine.NewRequest(topicOne),
		engine.NewParallel(
			engine.NewRequest(topicOne).With(StrMap{"merge-one": 1}),
			engine.NewRequest(topicTwo).With(StrMap{"merge-two": 1}),
			engine.NewRequest(topicThree).With(StrMap{"merge-three": 1}),
		),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	if resp.Cap() != 3 {
		t.Error("Must capture only 1 output.")
		return
	}

	out := []*Message{}
	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out = append(out, msg)
	}

	if len(out) != 3 {
		t.Error("Must have captured 3 outputs.")
	}

	expected := StrMap{}
	out[0].GetData(&expected)

	if _, ok := expected["input"]; !ok {
		t.Error("Must have key: input in final output")
	}
}

func TestPipeParallel(t *testing.T) {
	c := engine.NewPipe(
		engine.NewRequest(topicOne),
		engine.NewParallel(
			engine.NewRequest(topicOne),
			engine.NewRequest(topicTwo),
			engine.NewRequest(topicThree),
		),
	)

	resp, err := c.Execute(makeMessage(StrMap{"input": 1}))
	if err != nil {
		t.Error("Should not have raised Error")
		return
	}

	if resp.Cap() != 3 {
		t.Error("Must capture 3 outputs.")
	}

	out := []*Message{}
	for msg := resp.Next(); msg != nil; msg = resp.Next() {
		out = append(out, msg)
	}

	if len(out) != 3 {
		t.Error("Must have captured 3 outputs.")
	}

	expected := StrMap{}
	out[0].GetData(&expected)

	if _, ok := expected["ack-one"]; !ok {
		t.Error("Must have key: ack-one in final output")
	}
}
