package sketch

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type lyrics struct {
	column *column
	lyrics   *items.LyricsTable
}

func (c *lyrics) applyLyrics(evts []*items.Event) ([]*items.Event, error) {
	cc := c.lyrics

	idx := strings.Index(cc.Name, ".")
	if idx == -1 {
		return nil, fmt.Errorf("invalid lyrics call name %q (missing dot)", cc.Name)
	}

	if idx >= len(cc.Name)-1 {
		return nil, fmt.Errorf("invalid lyrics call name %q", cc.Name)
	}

	l, err := c.column.sketch.Score.Lyric(cc.Name, cc.Slice[0], cc.Slice[1])
	if err != nil {
		return nil, err
	}
	return applyLyrics(evts, l), nil
}
