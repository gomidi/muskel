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

func (t *Sketch) writeDataLine(f Formatter, line []string) (err error) {
	var s strings.Builder

	if len(line[0]) > 0 && (line[0][0] == '#' || line[0][0] == '$' || line[0][0] == '[' || line[0][0] == '*') {
		return t.Table.writeLine(f, " "+line[0])
	}
	s.WriteString(t.Table.Pad(0, "  "+line[0]) + t.Table.separator())

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
