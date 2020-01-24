package items

type rest struct{}

var Rest = rest{}

func (h rest) Parse(data string, posIn32th uint) (err error) {
	return nil
}

func (r rest) String() string {
	return "*"
}

func (r rest) Dup() Item {
	return Rest
}

var _ Item = Rest
