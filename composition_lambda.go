package gilmour

//Constructor for HashComposition
func (g *Gilmour) NewLambda(s func(*Message) (*Message, error)) *LambdaComposition {
	fc := &LambdaComposition{seed: s}
	return fc
}

// Standalone Merge transformer.
// Used if you want to transform the output of the previous command before
// seeding it to the next command. Requires to be seeded with an interface
// which will be applied to the previous command's output Message only if the
// type is convertible. In case of a failure it shall raise an Error.
type LambdaComposition struct {
	seed func(*Message) (*Message, error)
}

func (hc *LambdaComposition) Execute(m *Message) (*Response, error) {
	resp := newResponse(1)
	msg, err := hc.seed(m)
	resp.write(msg)
	return resp, err
}
