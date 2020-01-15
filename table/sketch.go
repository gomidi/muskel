package table

import (
	"strings"
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

func (s *Sketch) addCols() {
	for _, name := range s.Table.Cols() {
		s.sketch.AddColumn(name)
	}
}

func (s *Sketch) ParseLine(line string) error {
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
		return "  " + pos
	case ';':
		return "   " + pos
	default:
		return pos
	}
}

func (t *Sketch) writeDataLine(f Formatter, line []string) (err error) {
	var s strings.Builder

	if len(line[0]) > 0 && (line[0][0] == '#' || line[0][0] == '$' || line[0][0] == '[' || line[0][0] == '*') {
		var first, last string
		first = " " + line[0]
		if idx := strings.Index(line[0], "##"); idx > 0 {
			var ff formatLine
			t.Table.writeFirstLine(&ff)
			last = string(ff)
			idx2 := strings.Index(last, "|")
			last = "## " + last[idx2:]
			first = " " + line[0][:idx]
			//t.Table.Pad(0, first)
			first = Pad(first, t.Table.colWidths[0]-2)
			//fmt.Printf("first: %q last: %q\n", first, last)
		}
		return t.Table.writeLine(f, first+last)
	}
	s.WriteString(t.Table.Pad(0, "    "+formatPosition(line[0])) + t.Table.separator())

	for i, _ := range t.cols {
		col := ""
		if i+1 < len(line) {
			col = line[i+1]
		}
		s.WriteString(t.Table.Pad(i+1, col) + t.Table.separator())
	}

	if ll := len(line); ll > len(t.cols)+1 {
		s.WriteString(line[ll-1])
	}
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
