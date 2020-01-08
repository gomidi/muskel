package sketch

import "gitlab.com/gomidi/muskel/items"

type Event struct {
	Item     items.Item
	Position uint
}
