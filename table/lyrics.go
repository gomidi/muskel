package table

import (
	"fmt"
	"strings"
)

type Lyrics struct {
	*Table
}

func NewLyrics(name string, lineNo int, sc Score) *Lyrics {
	t := &Lyrics{
		Table: NewTable(name, lineNo, sc),
	}
	return t
}

func (t *Lyrics) Finish() (err error) {
	l := map[string][]string{}
	var part string
	for i, line := range t.Table.Data {
		if strings.TrimSpace(line[1]) != "" {
			part = strings.TrimSpace(line[1])
		}

		if part == "" {
			return fmt.Errorf("lyrics can't be added to empty part in line %v", t.lineNo+i)
		}
		l[t.name+"."+part] = append(l[t.name+"."+part], strings.TrimSpace(line[0]))
	}

	t.Score.AddLyrics(l)
	return nil
}
