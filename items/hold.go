package items

type hold struct{}

var Hold = hold{}
var _ Item = Hold

func (h hold) Parse(data string, posIn32th uint) (err error) {
	return nil
}

func (h hold) Dup() Item {
	return h
}

func (r hold) String() string {
	return ":"
}
