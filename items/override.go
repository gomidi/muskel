package items

import "fmt"

type Override struct {
	Item
}

func (o *Override) Dup() Item {
	return &Override{
		Item: o.Item.Dup(),
	}
}

func (o *Override) Parse(data string, posIn32th uint) error {
	if len(data) > 0 {
		it, err := Parse(data[1:], posIn32th)
		if err == nil {
			o.Item = it
			return nil
		}
		return err
		//return o.Item.Parse(data[1:], posIn32th)
	}
	return nil
}

var _ UnrollGetter = &Override{}

func (c *Override) GetES(p Columner, ev *Event, start, endPos uint) (mixed []*EventStream, err error) {
	var posEv *EventStream
	switch vv := c.Item.(type) {
	case *Pattern:
		posEv, err = vv.GetOverrideEventStream(p, start, endPos)
	case *Token:
		tk := c.Item.(*Token)
		posEv, err = tk.GetOverrideEventStream(p, start, start+1)
	case *Scale:
		return nil, fmt.Errorf("item type not allowed in override: %#v", vv)
	default:
		nev := ev.Dup()
		nev.Item = vv
		nev.Position = start
		posEv = NewEventStream(nev.Position, nev.Position+1, true, nev)
		posEv.IsOverride = true
	}
	if err != nil {
		return nil, err
	}
	return []*EventStream{posEv}, nil
}

var _ Item = &Override{}
