package items

import (
	"fmt"
)

type GlideStart bool

func (g GlideStart) Dup() Item {
	return &g
}

func (g GlideStart) String() string {
	if g {
		return "~~"
	}
	return "~"
}

func (g *GlideStart) Parse(data string, posIn32th uint) error {
	data = data[1:]
	if data == "" {
		*g = GlideLinear
		return nil
	}

	if data == "~" {
		*g = GlideExponential
		return nil
	}
	return fmt.Errorf("invalid glissando: \"~%s\"", data)
}

var GlideLinear = GlideStart(false)
var GlideExponential = GlideStart(true)
