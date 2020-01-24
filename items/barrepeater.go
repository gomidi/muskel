package items

import (
	"fmt"
	"strconv"
	"strings"
)

type BarRepeater struct {
	LastN    int
	OnlyOnce bool
}

func (b BarRepeater) Dup() Item {
	return &b
}

func (b BarRepeater) String() string {
	if b.OnlyOnce && b.LastN < 2 {
		return "./."
	}

	if b.LastN < 2 {
		return "..."
	}

	return fmt.Sprintf(".%v.", b.LastN)
}

func (b *BarRepeater) Parse(data string, posIn32th uint) (err error) {
	if data == ".." {
		b.OnlyOnce = false
		b.LastN = 1
		return nil
	}
	if data == "/." {
		b.OnlyOnce = true
		b.LastN = 1
		return nil
	}
	if strings.Index("123456789", data[0:1]) != -1 && data[1] == '.' {
		b.OnlyOnce = false
		b.LastN, err = strconv.Atoi(data[0:1])

		return
	}
	return fmt.Errorf("unsupported syntax: %q", "."+data)
}
