package items

import (
	"fmt"
	"strconv"
)

type Placeholder int

//var _ Item = &Placeholder{}

func (l *Placeholder) Parse(data string, posIn32th uint) (err error) {
	data = data[1:]

	i, err := strconv.Atoi(data)

	if err != nil {
		return err
	}

	*l = Placeholder(i)
	return nil
}

func (l Placeholder) String() string {
	return fmt.Sprintf("#%v", int(l))
}

func (v Placeholder) Dup() Item {
	return &v
}
