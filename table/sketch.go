package table

import (
	"sort"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type Sketch struct {
	*Table
	colsAdded bool
	sketch    sketch
}

func NewSketch(name string, lineNo int, sc Score) *Sketch {
	t := &Sketch{Table: NewTable(name, lineNo, sc)}
	t.sketch = sc.AddSketch(name)
	return t
}

func barLiner(data [][]string) (bars []map[string][]string) {
	currentBar := map[string][]string{}

	for _, d := range data {
		//	fmt.Printf("data: %#v\n", d[:len(d)-1])
		if len(d) > 0 {
			ch := strings.TrimSpace(d[0])
			if len(ch) > 0 && ch[0] == '#' {
				bars = append(bars, currentBar)
				currentBar = map[string][]string{
					"#": d[:len(d)-1],
				}
				continue
			}

			if len(ch) > 0 && len(d) > 1 {
				//bars = append(bars, currentBar)
				currentBar[strings.TrimSpace(d[0])] = d[1 : len(d)-1]
				continue
			}
		}
	}

	bars = append(bars, currentBar)
	return
}

type barPos struct {
	pos  float64
	vals []string
}

func convertPos(s string) float64 {
	_, pos, err := items.PositionTo32th(0, s)

	if err != nil {
		return 0
	}

	return float64(pos) / float64(32)
}

type barpositions []barPos

func (b barpositions) Len() int {
	return len(b)
}

func (bp barpositions) Swap(a, b int) {
	bp[a], bp[b] = bp[b], bp[a]
}

func (bp barpositions) Less(a, b int) bool {
	return bp[a].pos < bp[b].pos
}

func getSortedBarData(m map[string][]string) (lines [][]string) {
	if len(m) == 0 {
		return
	}
	bar := m["#"]

	lines = append(lines, bar)

	var bposes barpositions
	var posName = map[float64]string{}

	for k, v := range m {
		if k == "#" {
			continue
		}

		pos := convertPos(k)
		posName[pos] = k

		//	fmt.Printf("%q => %0.2f\n", k, pos)

		bp := barPos{
			pos:  pos,
			vals: v,
		}

		bposes = append(bposes, bp)
	}

	sort.Sort(bposes)

	for _, bp := range bposes {
		var vl = []string{posName[bp.pos]}
		vl = append(vl, bp.vals...)
		lines = append(lines, vl)
	}

	return
}

func (s *Sketch) Merge(o *Sketch) {
	l := len(s.cols)
	for i, c := range o.cols {
		s.cols = append(s.cols, c)
		if o.skipCols[i+1] {
			s.skipCols[l+i+1] = true
		}
	}

	origBars := barLiner(s.Data)
	otherBars := barLiner(o.Data)

	if len(origBars) < len(otherBars) {
		for _, b := range otherBars[len(origBars):] {

			m := map[string][]string{
				"#": b["#"],
			}

			origBars = append(origBars, m)
		}
	}

	for i, m := range otherBars {
		for k, v := range m {
			if k == "#" {
				continue
			}

			vn := make([]string, l+len(o.cols))
			for ii, vv := range v {
				//if len(v) > 1 {
				vn[l+ii] = vv

			}

			//	fmt.Printf("v: %#v in bar %v, k: %q\n", v, i, k)
			//	fmt.Printf("vn: %#v in bar %v, k: %q\n", vn, i, k)
			if vs, has := origBars[i][k]; has {
				// TODO make shure, vs has len(cols) (no missing cols)
				vs = append(vs, vn[1:]...)
				origBars[i][k] = vs
			} else {
				origBars[i][k] = vn
			}
		}
	}

	//s.Data
	var data [][]string

	for _, m := range origBars {
		lines := getSortedBarData(m)

		for _, l := range lines {
			//fmt.Printf("dataline: %#v\n", l)
			data = append(data, l)
		}

	}

	s.Data = data

	/*
		for _, d := range o.Data {
			d
		}
	*/
}

func (s *Sketch) addCols() {
	s.sketch.SetRealColNum(len(s.Table.Cols()))
	for i, name := range s.Table.Cols() {
		cols := strings.Split(strings.TrimSpace(name), " ")

		if len(cols) > 1 {
			var subCols []string

			for _, cl := range cols {
				cl = strings.TrimSpace(cl)
				if cl == "" {
					continue
				}
				subCols = append(subCols, cl)
				s.sketch.AddColumn(cl)
			}

			s.sketch.SetGroupCol(i, subCols)
		} else {
			s.sketch.AddColumn(strings.TrimSpace(name))
		}
	}
}

func (s *Sketch) ParseLine(line string) error {
	//line = strings.TrimRight(line, "|")
	//fmt.Printf("table.Sketch.ParseLine(%q)\n", line)
	if err := s.Table.ParseLine(line); err != nil {
		//fmt.Printf("table.Sketch.ParseLine err: %v\n", err)
		return err
	}
	if !s.colsAdded {
		s.addCols()
		s.colsAdded = true
	}
	if len(s.Table.Data) > 0 {
		//	fmt.Printf("s.Table.Data = %#v\n", s.Table.Data)
		return s.sketch.ParseLine(s.Table.Data[len(s.Table.Data)-1])
	}
	return nil
}

func (p *Sketch) WriteTo(f Formatter) error {
	if err := p.Table.CalcColWidths(); err != nil {
		return err
	}
	p.Table.writeFirstLine(f)
	return p.writeDataLines(f)
}

func formatPosition(pos string) string {
	pos = strings.TrimSpace(pos)
	if len(pos) == 0 {
		return ""
	}

	switch pos[0] {
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return pos
	case '&':
		return " " + pos
	case '.':
		return " " + pos
	case ';':
		return " " + pos
	default:
		return pos
	}
}

func (t *Sketch) writeDataLine(f Formatter, line []string) (err error) {
	var s strings.Builder
	if len(line[0]) > 0 && (line[0][0] == '#' || line[0][0] == '\'' || line[0][0] == '=' || line[0][0] == '[' || line[0][0] == '*') {
		var first, last string
		var skipCols bool
		first = " " + line[0]
		if idx := strings.Index(line[0], "##"); idx > 0 {
			var ff formatLine
			t.Table._writeFirstLine(&ff, true)
			last = string(ff)
			idx2 := strings.Index(last, "|")
			last = "## " + last[idx2:]
			first = " " + line[0][:idx]
			//t.Table.Pad(0, first)
			first = Pad(first, t.Table.colWidths[0]-2)
			skipCols = true
			//fmt.Printf("first: %q last: %q\n", first, last)
		}

		s.WriteString(t.Table.separator() + t.Table.Pad(0, first+last))

		if !skipCols {
			for i, _ := range t.cols {
				s.WriteString(t.Table.separator() + t.Table.Pad(i+1, " "))
			}
			s.WriteString(t.Table.separator())
		}
		return t.Table.writeLine(f, s.String())
	}

	if t.Score != nil && t.Score.NoEmptyLines() && t.IsEmptyLine(line[1:]) {
		//fmt.Printf("line[1:]: %#v\n", line[1:])
		return
	}

	s.WriteString(t.Table.separator() + t.Table.Pad(0, "    "+formatPosition(line[0])))

	for i, _ := range t.cols {
		col := ""
		if i+1 < len(line) {
			col = line[i+1]
		}
		s.WriteString(t.Table.separator() + t.Table.Pad(i+1, col))
	}

	if ll := len(line); ll > len(t.cols)+1 {
		s.WriteString(line[ll-1])
	}

	s.WriteString(t.Table.separator())
	return t.Table.writeLine(f, s.String())
}

func (t *Sketch) writeDataLines(f Formatter) (err error) {
	for _, line := range t.Table.Data {
		err = t.writeDataLine(f, line)
		if err != nil {
			return
		}
	}
	return
}

func (s *Sketch) IsEmptyLine(data []string) bool {
	for _, d := range data {
		if strings.TrimSpace(d) != "" {
			return false
		}
	}

	return true
}
