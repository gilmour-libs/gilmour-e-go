package gilmour

import "testing"

func TestMergeHash(t *testing.T) {
	src := map[string]interface{}{"a": 1}
	merg := map[string]interface{}{"b": 1}
	err := compositionMerge(src, merg)
	if err != nil {
		t.Error("Should mrge cleanly")
	}

	if _, ok := src["b"]; !ok {
		t.Error("Should have had the key b merged.")
	}
}

func TestMergeMismatch(t *testing.T) {
	src := "a"
	merg := map[string]interface{}{"b": 1}
	err := compositionMerge(src, merg)
	if err == nil {
		t.Error("Should give merge conflict")
	}
}

func TestMergePointer(t *testing.T) {
	src := map[string]interface{}{"a": 1}
	merg := map[string]interface{}{"b": 1}
	err := compositionMerge(&src, merg)
	if err != nil {
		t.Error("Should mrge cleanly")
	}

	if _, ok := src["b"]; !ok {
		t.Error("Should have had the key b merged.")
	}
}

func TestMergeInterface(t *testing.T) {
	src := map[string]interface{}{"a": 1}
	merg := map[string]interface{}{"b": 1}
	err := compositionMerge(
		func() interface{} {
			return src
		}(),
		func() interface{} {
			return merg
		}(),
	)

	if err != nil {
		t.Error("Should mrge cleanly")
	}

	if _, ok := src["b"]; !ok {
		t.Error("Should have had the key b merged.")
	}
}

func TestMergeInterfacePointer(t *testing.T) {
	src := map[string]interface{}{"a": 1}
	merg := map[string]interface{}{"b": 1}
	err := compositionMerge(
		func() interface{} {
			return &src
		}(),
		func() interface{} {
			return merg
		}(),
	)

	if err != nil {
		t.Error("Should mrge cleanly")
	}

	if _, ok := src["b"]; !ok {
		t.Error("Should have had the key b merged.")
	}
}
