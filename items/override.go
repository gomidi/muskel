package items

type Override struct {
	Item
}

func (o *Override) Dup() Item {
	return &Override{
		Item: o.Item.Dup(),
	}
}

var _ Item = &Override{}
