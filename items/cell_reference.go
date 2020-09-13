package items

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/reference"
)

type CellReference struct {
	*reference.Reference
	data   string
	params []string
}

func (c *CellReference) Dup() Item {
	i := &CellReference{
		data: c.data,
	}

	if c.Reference != nil {
		r := c.Reference.Dup()
		i.Reference = &r
	}
	return i
}

func (c *CellReference) GetItem(targetCol string) (Item, error) {
	ctx := reference.NewScoreColCtx("notimportant-file", "=SCORE", targetCol)

	err := c.Reference.Complete(ctx)

	if err != nil {
		return nil, err
	}

	switch c.Reference.Type {
	case reference.ScoreCol:
		var pt Pattern
		pt.Params = c.params
		pt.IncludeFile = c.Reference.File
		pt.Name = c.Reference.Table + "." + c.Reference.Col
		// TODO what to do about the other properties?? dynamic add etc
		return &pt, nil
		// include a col
	case reference.ScoreColPart:
		var pt Pattern
		pt.Params = c.params
		pt.IncludeFile = c.Reference.File
		pt.Name = c.Reference.Table + "." + c.Reference.Col
		pt.Part = c.Reference.Part
		// TODO what to do about the other properties?? dynamic add etc
		return &pt, nil
		// include a part of a col
	case reference.ShortCutCell:
		// include an external token
		var t Token
		t.IncludeFile = c.Reference.File
		if c.Reference.Col != "" {
			t.Name = c.Reference.Table + "." + c.Reference.Row + "." + c.Reference.Col
		} else {
			t.Name = c.Reference.Table + "." + c.Reference.Row
		}
		t.itemGroupModifier.Params = c.params
		// TODO what to do about the other properties of itemGroupModifier?? dynamic add, slicing with [] etc
		return &t, nil
	default:
		return nil, fmt.Errorf("invalid reference type %s in cell reference", c.Reference.Type.String())
	}
}

func (c *CellReference) Parse(data string, posIn32th uint) error {
	if idx := strings.Index(data, "("); idx > -1 {
		endIdx := strings.Index(data, ")")
		if endIdx == -1 {
			return fmt.Errorf("invalid cell reference: %q", data)
		}
		c.params = strings.Split(data[idx+1:endIdx], ",")
		c.data = data[:idx]
	} else {
		c.data = data
	}

	r, err := reference.Parse(data)
	if err != nil {
		return err
	}

	c.Reference = r
	return nil
}

func (c *CellReference) String() string {
	return c.data
}
