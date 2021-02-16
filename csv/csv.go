package csv

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
	// "gitlab.com/gomidi/muskel/score"
	// "gitlab.com/gomidi/muskel/table"
)

func Read(rd io.Reader, seperator rune) (muskelFormat string, err error) {
	r := csv.NewReader(rd)
	//r.Comma = ';'
	r.Comma = seperator
	r.FieldsPerRecord = -1
	//r.Comment = '#'

	var reader Reader

	for {

		rec, err := r.Read()
		if err == io.EOF {
			break
		}

		//fmt.Printf("record: %#v\n", record)

		reader.WriteRecord(rec)

	}

	return reader.String(), nil
}

type Reader struct {
	bf          bytes.Buffer
	insideTable bool
}

func (r *Reader) WriteString(s string, insideTable bool) {
	r.insideTable = insideTable
	r.bf.WriteString(s)
}

func (r *Reader) WriteRecord(rec []string) {

	var record []string

	var wasEmpty bool

	for _, re := range rec {
		if len(re) == 0 {
			if !wasEmpty {
				record = append(record, "")
			}
			wasEmpty = true
			continue
		} else {
			wasEmpty = false
		}

		if len(re) > 1 && re[0] == '\'' {
			record = append(record, re[1:])
		} else {
			record = append(record, re)
		}
	}

	if len(record) == 0 {
		r.bf.WriteString("\n")
		r.insideTable = false
		return
	}

	if record[0] == "" {
		r.bf.WriteString("\n")
		r.insideTable = false
		return
	}

	writeSimple := false
	switch {
	// a comment of some kind
	case record[0][0] == '/':
		if len(record[0]) > 1 {
			switch record[0][1] {
			case '*': // multiline comment start
				writeSimple = true
			case '/': // single line comment start
				writeSimple = true
			default:
			}
		}

	// multiline comment end
	case record[0] == "*/":
		writeSimple = true
	// we are in a score table
	case record[0][0] == '=':
		r.insideTable = true
	/*
		case record[0][0] == '\'':
			if len(record[0]) > 1 {
				if record[0][1] == '=' {
					insideTable = true
				}
				record[0] = record[0][1:]
			}
	*/
	// we have shortcuts
	case record[0][0] == '.':
		r.insideTable = true
	// we have lyrics
	case record[0][0] == '@':
		r.insideTable = true
	case record[0] == "TRACK":
		r.insideTable = true
	case record[0] == "SCALE":
		r.insideTable = true
	case record[0] == "TUNING":
		r.insideTable = true
	case record[0] == "TIMBRE":
		r.insideTable = true
	case record[0] == "FILTER":
		r.insideTable = true
	case record[0] == "PROPERTY":
	default:
		//fmt.Printf("inside table %v\n", insideTable)
		if !r.insideTable {
			writeSimple = true
		}
	}

	var str string

	if writeSimple {
		str = strings.Join(record, " ")
	} else {
		str = "|" + strings.Join(record, " | ") + " |"
	}

	//fmt.Printf("%q\n", str)

	r.bf.WriteString(str + "\n")
}

func (r *Reader) String() string {
	return r.bf.String()
}

/*
func Write(file string) {

	records := [][]string{
		{"first_name", "last_name", "occupation"},
		{"John", "Doe", "gardener"},
		{"Lucy", "Smith", "teacher"},
		{"Brian", "Bethamy", "programmer"},
	}

	f, err := os.Create("users.csv")
	defer f.Close()

	if err != nil {

		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	w.Comma = ';'
	defer w.Flush()

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}
}
*/
