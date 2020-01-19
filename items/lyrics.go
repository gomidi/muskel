package items

import (
	"fmt"
	"strings"
)

type Lyric struct {
	Text     string
	PosShift int // 0 = no, 1 = laidback, -1 = ahead of time
}

func (l *Lyric) Parse(data string, posIn32th uint) (err error) {
	t := strings.Trim(data, `"`)
	l.Text = t
	return nil
}

func (l *Lyric) Dup() Item {
	return &Lyric{Text: l.Text, PosShift: l.PosShift}
}

func (l *Lyric) String() string {
	s := fmt.Sprintf("%q", l.Text)
	switch l.PosShift {
	case -1:
		return s + "<"
	case 1:
		return s + ">"
	default:
		return s
	}
}

func (v *Lyric) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	wr.Lyric(v.Text)
	return
}
