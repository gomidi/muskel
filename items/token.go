package items

import (
	"fmt"
)

type Token struct {
	Name string
	itemGroupModifier
}

func (t *Token) newSketchToken(column Columner) *sketchToken {
	return &sketchToken{
		column: column,
		token:  t,
	}
}

func (t *Token) GetEventStream(column Columner, start, until uint) (*EventStream, error) {
	pc := t.newSketchToken(column)
	return pc.getEventStream(start, until)
}

func (t *Token) GetOverrideEventStream(column Columner, start, until uint) (*EventStream, error) {
	pc := t.newSketchToken(column)
	return pc.getOverrideEventStream(start, until)
}

var _ UnrollGetter = &Token{}

func (c *Token) GetES(p Columner, ev *Event, start, endPos uint) (mixed []*EventStream, err error) {
	// TODO prevent endless loops from templates calling each other like col1 -> col2 -> col1 by keeping a stack of template calls
	// and checking them for duplicates (the stack may as well be a map[string]bool; makes is easier; we need the complete names in there
	posEv, err := c.GetEventStream(p, start, endPos)
	if err != nil {
		return nil, err
	}
	return []*EventStream{posEv}, nil
}

func (t *Token) Parse(data string, positionIn32th uint) (err error) {
	data = t.parseExploded(data)
	mt := regTokenCallNameDyn.FindStringSubmatch(data)
	t.Name = mt[1]
	if len(t.Name) == 0 || t.Name[0] != '.' {
		return fmt.Errorf("invalid token name: %q", t.Name)
	}

	err = t.itemGroupModifier.parse(mt[2])

	if err != nil {
		return fmt.Errorf("error while parsing token %q: %s", data, err.Error())
	}

	return nil
}

func (t Token) Dup() Item {
	return &t
}

func (t *Token) String() string {
	return t.Name + t.itemGroupModifier.string()
}

/*
func (t *Token) SetDynamic(dyn string) {
	// TODO resolv and then set it
	//nt.Dynamic = AddDynamic(nt.Dynamic, dyn)
}
*/
