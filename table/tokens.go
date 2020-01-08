package table

import (
	"fmt"
)

type Tokens struct {
	*Table
}

func NewTokens(name string, lineNo int, sc Score) *Tokens {
	if name[0] == '=' {
		panic(fmt.Sprintf("invalid token table name %q", name))
	}

	t := &Tokens{
		Table: NewTable(name, lineNo, sc),
	}
	//sc.filters = t
	return t
}

func (t *Tokens) Finish() (err error) {
	for _, line := range t.Table.Data {
		if len(line) < 2 {
			return fmt.Errorf("ivalid token table %q need at least 2 columns", t.name)
		}
		t.Table.Score.AddToken(line[0], line[1])
		t.Table.Score.AddToken(t.name+"."+line[0], line[1])
	}

	t.cols = nil

	return nil
}
