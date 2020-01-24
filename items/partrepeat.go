package items

import (
	"fmt"
	"strconv"
	"strings"
)

type PartRepeat struct {
	Part      string
	Repeat    uint
	SyncFirst bool
}

var _ Item = &PartRepeat{}

func (pr *PartRepeat) Dup() Item {
	return &PartRepeat{
		Part:      pr.Part,
		Repeat:    pr.Repeat,
		SyncFirst: pr.SyncFirst,
	}
}

func (pr *PartRepeat) String() string {
	var bd strings.Builder
	bd.WriteString("[" + pr.Part + "]")
	if pr.Repeat > 1 {
		bd.WriteString("%" + fmt.Sprintf("%v", pr.Repeat))
	}
	return bd.String()
}

// ntuple has the form {c,e,d}>
func (pr *PartRepeat) Parse(data string, posIn32th uint) (err error) {

	orig := data

	idx := strings.LastIndex(data, "]")

	if idx <= 0 {
		return fmt.Errorf("invalid part repeat: %#v", "["+orig)
	}

	pr.Repeat = 1

	if idx < len(data)+2 {
		dt := data[idx+1:]
		if idx2 := strings.LastIndex(dt, "%"); idx2 > -1 {
			numstr := dt[idx2:]
			var repeat int
			repeat, err = strconv.Atoi(numstr)
			if err != nil {
				err = fmt.Errorf("invalid repeat syntax: %q is not a number", numstr)
				return
			}

			if repeat <= 0 {
				err = fmt.Errorf("invalid repeat number: %v (must be > 0)", repeat)
				return
			}

			pr.Repeat = uint(repeat)
		}
	}

	data = strings.TrimSpace(data[:idx])

	if data[0] == '!' {
		pr.SyncFirst = true
		data = data[1:]
	}
	pr.Part = data
	return nil
}
