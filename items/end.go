package items

type end struct{}

var End = end{}

func (h end) Parse(data string, posIn32th uint) (err error) {
	return nil
}

func (r end) String() string {
	return "_"
}

func (r end) Dup() Item {
	return End
}

var _ Item = End
