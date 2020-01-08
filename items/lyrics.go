package items

import (
	"fmt"
	"strings"
)

type Lyric string

func (l *Lyric) Parse(data string, posIn32th uint) (err error) {
	t := strings.Trim(data, `"`)
	*l = Lyric(t)
	return nil
}

func (l Lyric) Text() string {
	return string(l)
}

func (l Lyric) Dup() Item {
	return &l
}

func (l Lyric) String() string {
	return fmt.Sprintf("%q", string(l))
}

func (v Lyric) WriteMIDI(wr SMFWriter) (addedNotes []uint8) {
	wr.Lyric(v.Text())
	return
}
