package items

import (
	"fmt"
)

type Token struct {
	Name string
	itemGroupModifier
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

func (p *Token) AddDynamic(orig string) (nu string) {
	//fmt.Printf("addDynamic %q to %q\n", p.DynamicAdd, orig)
	if orig == "=" {
		return orig
	}

	if p.DynamicAdd == "" {
		return orig
	}

	if p.DynamicAdd == "=" {
		return p.DynamicAdd
	}

	return reduceDynamic(orig + p.DynamicAdd)
}
