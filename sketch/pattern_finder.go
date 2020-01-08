package sketch

import (
	"fmt"
	"strings"
)

type patternFinder struct {
	fromSketch *Sketch
}

func (r *patternFinder) FindPattern(callingCol string, callName string) (s *Sketch, colName string, err error) {
	/*
	  =.a ist das pattern im gleichen sketch in der spalte a, alle patterns des gleichen sketches müssen über =. angesprochen werden (und nicht über den Namen des sketches)
	  =~.a ist das pattern in der spalte a im sketch der heißt, wie die aktuelle spalte
	  =.~ ist nicht zulässig
	  =aa.~ ist das pattern im sketch =aa in der spalte, die den gleichen namen trägt wie die aktuelle spalte
	  =aa ist das pattern im sketch =aa, welches in der ersten spalte ist
	  die letzte spalte, die als Namen mit "!" beginnt, wird als erste Spalte behandelt, gibt es keine solche spalte, wird die
	  tatsächlich erste spalte genommen
	*/

	callName = strings.TrimSpace(callName)
	callingCol = strings.TrimSpace(callingCol)

	if callName[0] != '=' {
		err = fmt.Errorf("can't resolve pattern name %q: name must start with =", callName)
		return
	}

	if callName == "=.~" {
		err = fmt.Errorf("can't resolve pattern name %q: name must not be =.~ (self reference)", callName)
		return
	}

	_x := strings.Index(callName, ".")
	sname := callName
	var cname string

	if _x > 0 {
		sname = callName[:_x]
		if len(callName) > _x {
			cname = callName[_x+1:]
		}
	}

	if sname == r.fromSketch.Name {
		err = fmt.Errorf("can't resolve pattern name %q: same sketch must be referrenced via =.col", callName)
		return
	}

	sname = strings.ReplaceAll(sname, "~", callingCol)
	colName = strings.ReplaceAll(cname, "~", callingCol)

	if sname == "=" {
		if callingCol == colName {
			err = fmt.Errorf("can't resolve pattern name %q from column: %q (self reference)", callName, callingCol)
		}
		s = r.fromSketch
	} else {
		s, err = r.fromSketch.Score.GetSketch(sname)
		if err != nil {
			return
		}
	}

	if len(s.colOrder) == 0 {
		err = fmt.Errorf("can't resolve pattern name %q: sketch %q has no columns", callName, s.Name)
		return
	}

	if colName == "" {
		for i, col := range s.colOrder {
			if i == 0 || col[0] == '!' {
				colName = col
			}
		}
	}

	return
}
