package table

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type Scales struct {
	*Table
}

func NewScales(lineNo int, sc Score) *Scales {
	return &Scales{
		Table: NewTable("SCALE", lineNo, sc),
	}
}

func (s *Scales) Finish() (err error) {

	for i, col := range s.cols {
		var cdata []string
		for _, data := range s.Data {
			cdata = append(cdata, data[i+1])
		}

		m, err := s.getMode(col, cdata)
		if err != nil {
			return fmt.Errorf("invalid scale %q: %s", col, err.Error())
		}
		_ = m
		s.Score.AddScale(m)
	}
	return nil
}

func (s *Scales) getMode(name string, data []string) (m items.Mode, err error) {
	// `c",d",e",f",g",a",b"`,
	// a" ist der referenzpunkt
	var def items.ScaleDefinition
	err = def.Parse(name, strings.Join(data, ","))
	if err != nil {
		return nil, err
	}
	return &def, nil
}
