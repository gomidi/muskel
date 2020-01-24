package items

type Item interface {
	String() string
	Parse(data string, posIn32th uint) error
	Dup() Item
}

func ParseNote(data string) (nt Note, err error) {
	err = nt.Parse(data, 0)
	return
}
