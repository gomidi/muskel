package items

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

var _ Item = &Override{}
