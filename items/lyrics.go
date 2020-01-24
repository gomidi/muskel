package items

import (
	"fmt"
	"strings"
)

type Lyric struct {
	Text string
}

func (l *Lyric) Parse(data string, posIn32th uint) (err error) {
	t := strings.Trim(data, `"`)
	l.Text = t
	return nil
}

func (l *Lyric) Dup() Item {
	return &Lyric{Text: l.Text}
}

func (l *Lyric) String() string {
	s := fmt.Sprintf("%q", l.Text)
	return s
}
