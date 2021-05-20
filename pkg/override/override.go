package override

type Override struct {
	Handler  *interface{}
	Mock     interface{}
	Original interface{}
}

func (o Override) Apply() {
	o.Original = *o.Handler
	*o.Handler = o.Mock
}

func (o Override) Remove() {
	*o.Handler = o.Original
}

type Result struct {
	Output      []byte
	ErrorOutput []byte
	Err         error
}

type Action func() Result

func WithOverrides(action Action, overrides ...Override) Result {
	for _, o := range overrides {
		o.Apply()
	}
	res := action()
	for _, o := range overrides {
		o.Remove()
	}
	return res
}
