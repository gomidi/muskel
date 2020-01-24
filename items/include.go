package items

import (
	"fmt"
	"strings"
)

type Include struct {
	File        string
	Sketch      string
	Params      []string
	Position    uint
	Length32ths uint
}

func (i Include) Dup() Item {
	return i
}

func (i Include) String() string {
	var bf strings.Builder
	bf.WriteString("$include(" + fmt.Sprintf("%q", i.File))

	if i.Sketch != "" {
		bf.WriteString("," + fmt.Sprintf("%q", i.Sketch))
	}

	for _, p := range i.Params {
		bf.WriteString("," + p)
	}

	bf.WriteString(")")

	return bf.String()
}

func (i Include) Parse(data string, posIn32th uint) error {
	panic("don't call me")
}

var _ Item = Include{}
