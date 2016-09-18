package gilmour

import "testing"

func TestResponseBufferOverflow(t *testing.T) {
	x := newResponse(1)
	x.write(NewMessage())
	err := x.write(NewMessage())
	if err == nil {
		t.Error("Should complain about buffer overflow")
	}
}

func TestResponseNext(t *testing.T) {
	x := newResponse(2)
	x.write(NewMessage())

	err := x.write(NewMessage())
	if err != nil {
		t.Error("Should allow write twice")
	}

	recv := 0
	for y := x.Next(); y != nil; y = x.Next() {
		recv++
	}

	if recv != 2 {
		t.Error("Should have returned twice")
	}
}
