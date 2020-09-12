package table

import (
	"fmt"
	"strings"
)

type Table struct {
	printSlim       bool
	Score           Score
	name            string
	lineNo          int
	firstLineParsed bool
	cols            []string
	skipCols        map[int]bool
	Data            [][]string // lines -> cols
	colWidths       []int
}

func NewTable(name string, lineNo int, sc Score) *Table {
	//fmt.Printf("new table %q\n", name)
	return &Table{name: name, lineNo: lineNo, Score: sc, skipCols: map[int]bool{}}
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) AddCol(name string) {
	//fmt.Printf("table %q addCol %q\n", t.name, name)
	t.cols = append(t.cols, name)
}

func (t *Table) AddLine(data []string) {
	//fmt.Printf("table %q addLine %q\n", t.name, data)
	t.Data = append(t.Data, data)
}

func (t *Table) Cols() []string {
	return t.cols
}

func (t *Table) LineNo() int {
	return t.lineNo
}

func (i *Table) Pad(col int, s string) string {
	if i.printSlim {
		return Pad(s, i.colWidths[col])
	}
	return Pad(s, i.colWidths[col]+1)
}

func (t *Table) CalcColWidths() error {
	t.colWidths = make([]int, len(t.cols)+1)

	// set the first column (table name etc)
	err := t.calcColWidth(0)
	if err != nil {
		return err
	}

	for i := range t.cols {
		err = t.calcColWidth(i + 1)
		if err != nil {
			return err
		}
	}

	for _, line := range t.Data {
		for i, c := range line {
			if i == 0 && len(line) == 1 && len(line[0]) > 0 {
				switch line[0][0] {
				//case '$':
				//	continue
				case '\'':
					continue
				case '=':
					continue
				case '[':
					continue
				case '*':
					continue
				default:
				}
			}

			if idx := strings.Index(c, "##"); idx > 0 {
				c = c[:idx+2]
			} else {
				if idx2 := strings.Index(c, "//"); idx2 > 0 {
					c = c[:idx2]
				}
			}

			if i == 0 {
				c = c + " "
			}

			/*

			 */
			if ll := len(c); i < len(t.cols)+1 && ll > t.colWidths[i] {
				t.colWidths[i] = ll
			}
		}
	}

	return nil
}

func (t *Table) calcColWidth(col int) error {

	if col == 0 {
		t.colWidths[0] = LenUTF8(t.name)

		if t.colWidths[0] < 8 {
			t.colWidths[0] = 8
		}

		return nil
	}

	if col-1 > len(t.cols)-1 {
		return fmt.Errorf("no column title at index %v inside %v", col-1, t.cols)
	}

	t.colWidths[col] = LenUTF8(t.cols[col-1])

	if t.colWidths[col] < 1 {
		t.colWidths[col] = 1
	}

	return nil
}

func (t *Table) Finish() error {
	return nil
}

func (t *Table) tabs(line string) []string {
	c := strings.Split(line, "|")

	cc := make([]string, len(c))

	for i, _c := range c {
		cc[i] = strings.TrimSpace(_c)
	}

	return cc
}

func (t *Table) ParseLine(line string) error {
	//fmt.Printf("table: parseLine %q\n", line)
	if !t.firstLineParsed {
		tName := strings.TrimSpace(strings.Split(line, "|")[0])
		if !tableNameRegEx.MatchString(tName) {
			return fmt.Errorf("invalid syntax table name: %q", tName)
		}

		cols := tableHeader(line)
		for i, col := range cols {
			if strings.TrimSpace(col) == "-" {
				t.skipCols[i+1] = true
			} else {
				t.cols = append(t.cols, col)
			}
		}

		//fmt.Printf("cols: %v\n", t.cols)
		//t.cols
		t.firstLineParsed = true
		return nil
	}

	//fmt.Printf("add to data: %v\n", t.tabs(line))
	_tabs := t.tabs(line)
	var tabs []string

	for i, tab := range _tabs {
		if !t.skipCols[i] {
			tabs = append(tabs, tab)
		}
	}

	t.Data = append(t.Data, tabs)
	return nil
}

func (t *Table) writeLine(f Formatter, s string) error {
	return f.WriteLine(strings.TrimRight(s, " "))
}

func (t *Table) writeFirstLine(f Formatter) {
	//fmt.Printf("write first line of table: %q\n", t.name)
	var s strings.Builder
	s.WriteString(t.Pad(0, t.name) + t.separator())

	for i, c := range t.cols {
		s.WriteString(t.Pad(i+1, c) + t.separator())
	}

	t.writeLine(f, s.String())
}

func (t *Table) separator() string {
	if t.printSlim {
		return "|"
	}

	return "| "
}

func (t *Table) writeDataLine(f Formatter, line []string) (err error) {
	//fmt.Printf("write data line: %v\n", line)
	var s strings.Builder

	s.WriteString(t.Pad(0, " "+line[0]) + t.separator())

	for i, _ := range t.cols {
		col := ""

		if i+1 < len(line) {
			col = line[i+1]
		}

		s.WriteString(t.Pad(i+1, col) + t.separator())
	}

	if ll := len(line); ll > len(t.cols)+1 {
		s.WriteString(line[ll-1])
	}

	return t.writeLine(f, s.String())
}

func (t *Table) writeDataLines(f Formatter) (err error) {
	for _, line := range t.Data {
		err = t.writeDataLine(f, line)
		if err != nil {
			return
		}
	}

	return
}

func (t *Table) WriteTo(f Formatter) error {
	err := t.CalcColWidths()

	if err != nil {
		return err
	}

	t.writeFirstLine(f)

	return t.writeDataLines(f)
}
