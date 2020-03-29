package items

import (
	"fmt"
	"strconv"
	"strings"
)

type EuclideanRhythm struct {
	StartPos32th uint
	Notes        int
	Positions    int
	Duration32th int
	Sketch       [][2]string
	LastPos      uint
}

func (e *EuclideanRhythm) Parse(startPos32th uint, params ...string) error {
	if len(params) != 3 {
		return fmt.Errorf("EuclideanRhythm must have 3 params")
	}

	e.StartPos32th = startPos32th

	i, err := strconv.Atoi(params[0])

	if err != nil {
		return fmt.Errorf("number of notes must be an integer")
	}

	e.Notes = i

	i, err = strconv.Atoi(params[1])

	if err != nil {
		return fmt.Errorf("number of positions must be an integer")
	}

	e.Positions = i

	switch len(params[2]) {
	case 1:
		switch params[2] {
		case "&":
			e.Duration32th = 4
		case ",":
			e.Duration32th = 2
		case ";":
			e.Duration32th = 1
		default:
			qn, err := strconv.Atoi(params[2])
			if err != nil {
				return fmt.Errorf("invalid duration: %#v", params[3])
			}

			e.Duration32th = 8 * qn

		}
	case 0:
		return fmt.Errorf("invalid duration")
	default:
		if idx := strings.Index(params[2], "."); idx > 0 {
			f, err := strconv.ParseFloat(params[2], 64)
			if err != nil {
				return fmt.Errorf("invalid duration: %#v", params[2])
			}

			switch f {
			case 0.5:
				e.Duration32th = 4
			case 0.25:
				e.Duration32th = 2
			case 0.125:
				e.Duration32th = 1
			default:
				return fmt.Errorf("invalid duration: %#v", params[2])
			}
		} else {
			qn, err := strconv.Atoi(params[2])
			if err != nil {
				return fmt.Errorf("invalid duration: %#v", params[2])
			}

			e.Duration32th = 8 * qn
		}
	}

	line := euclideanRhythm(e.Notes, e.Positions)
	e.Sketch = append(e.Sketch, [2]string{"#", ""})
	var currentBar uint

	for i, set := range line {
		ap := e.StartPos32th + uint(i*e.Duration32th) - (currentBar * 32)

		if ap >= 32 {
			currentBar++
			e.Sketch = append(e.Sketch, [2]string{"#", ""})
			ap -= 32
		}
		pos := Pos32thToString(ap)
		e.LastPos = ap
		var val string
		if set {
			val = "#1"
		} else {
			val = "#2"
		}

		e.Sketch = append(e.Sketch, [2]string{pos, val})
	}

	return nil
}

func euclideanRhythm(onNotes, totalNotes int) []bool {
	groups := [][]bool{}

	for i := 0; i < totalNotes; i++ {
		groups = append(groups, []bool{i < onNotes})
	}

	for l := len(groups) - 1; l > 0; l = len(groups) - 1 {
		var start = 0
		var first = groups[0]
		for start < l && compareArrays(first, groups[start]) {
			start++
		}
		if start == l {
			break
		}

		var end = l
		var last = groups[l]

		for end > 0 && compareArrays(last, groups[end]) {
			end--
		}
		if end == 0 {
			break
		}

		var count = min(start, l-end)

		groups1 := slice(0, count, groups)
		groups2 := mapp(groups1, func(gr []bool, i int) []bool {
			return append(gr, groups[l-i]...)
		})

		groups = append(groups2, slice(count, -count, groups)...)
	}

	var res []bool

	for _, gr := range groups {
		res = append(res, gr...)
	}

	return res
}
