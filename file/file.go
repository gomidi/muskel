package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/gomidi/muskel/table"
)

type File struct {
	Score score
	Input *bufio.Scanner
	Parts []Part

	name                    string
	dir                     string
	file                    *os.File
	multiComments           map[int]*MultiLineComment
	commentLines            map[int]string
	currentLine             int
	currentTable            tabler
	currentMultiLineComment *MultiLineComment
	output                  io.Writer
}

func (f *File) Name() string {
	return filepath.Join(f.dir, f.name)
}

func (f *File) writeComment() (bool, error) {
	if cmt, has := f.commentLines[f.currentLine]; has {
		err := f.println(cmt)
		if err != nil {
			return false, err
		}

		//fmt.Println("writing comment line")

		return true, nil
	}

	for start, mcomment := range f.multiComments {
		if mcomment.Contains(f.currentLine) {
			//fmt.Printf("writing multicomment line: %v\n", f.currentLine)
			if f.currentLine == mcomment.Start() {
				err := f.println("/*")
				return true, err
			}

			if f.currentLine == mcomment.End() {
				err := f.println("*/")
				return false, err
			}
			err := f.println(mcomment.lines[f.currentLine-start-1])
			return true, err
		}
	}

	return false, nil
}

func (f *File) WriteLine(line string) (err error) {
	hasComment := true
	for hasComment {
		hasComment, err = f.writeComment()
		if err != nil {
			return err
		}
	}
	return f.println(line)
}

func (f *File) println(line string) error {
	_, err := fmt.Fprintln(f.output, line)
	//fmt.Fprintf(os.Stderr, "printing line %v: %q\n", f.currentLine, line)
	f.currentLine++
	return err
}

func (f *File) Format() error {
	fpath := filepath.Join(f.dir, f.name)
	os.Remove(fpath)

	fl, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.WriteTo(fl)

}

func (f *File) WriteTo(output io.Writer) (err error) {
	f.output = output
	/*
		var parts []string

		for _, part := range f.Parts {
			//fmt.Printf("printing part %T\n", part)
			parts = append(parts, fmt.Sprintf("%T %v", part, part.LineNo()))
		}
	*/

	//	fmt.Printf("parts: %v\n", parts)

	/*
		fmt.Println("mcomments: ")

		for no, ln := range f.multiComments {
			fmt.Printf(" %v/%v->%q\n", no, ln.lineNo, ln.lines)
		}

		fmt.Println(" ")
	*/
	//fmt.Printf("comment_lines: %v\n", f.commentLines)

	if len(f.Parts) == 0 {
		return fmt.Errorf("no parts defined")
	}

	f.currentLine = 0

	firstPart := f.Parts[0]
	start := firstPart.LineNo()
	//fmt.Printf("start: %v\n", start)
	hasComment := true
	for f.currentLine < start && hasComment {
		hasComment, err = f.writeComment()
		if err != nil {
			return fmt.Errorf(
				"ERROR when writing line %v: %v",
				f.currentLine, err,
			)
		}

	}

	for _, part := range f.Parts {
		err = part.WriteTo(f)
		if err != nil {
			return fmt.Errorf(
				"ERROR when writing line %v: %v",
				f.currentLine, err,
			)
		}
	}

	hasComment = true

	for hasComment {
		hasComment, err = f.writeComment()
		if err != nil {
			return fmt.Errorf(
				"ERROR when writing line %v: %v",
				f.currentLine, err,
			)
		}
	}

	return nil
}

func (f *File) findPart(line string) interface{} {
	if len(line) > 1 && line[:2] == "/*" {
		if f.currentMultiLineComment != nil {
			return nil
		}
		return &MultiLineComment{lineNo: f.currentLine}
	}

	if f.currentMultiLineComment != nil {
		return f.currentMultiLineComment
	}

	/*
		if len(line) > 1 && line[0] == '@' {
			return &Property{lineNo: f.currentLine, score: f.Score}
		}
	*/

	if len(line) > 1 && line[:2] == "//" {
		return line
	}

	if strings.TrimSpace(line) == "" {
		return EmptyLine(f.currentLine)
	}

	if f.currentTable != nil {
		return f.currentTable
	}

	if includeRegExp.MatchString(strings.ToLower(line)) {
		return &Include{
			lineNo: f.currentLine,
			score:  f.Score,
		}
	}
	
	if embedRegExp.MatchString(strings.ToLower(line)) {
		return &Embed{
			lineNo: f.currentLine,
			score:  f.Score,
		}
	}

	idx := strings.Index(line, "|")
	name := strings.TrimSpace(line)
	if idx > 1 {
		name = strings.TrimSpace(line[:idx])
	}
	//rest := strings.TrimSpace(line[idx+1:])

	//switch {
	//case strings.Contains(rest, "|"):
	switch name {
	case "TRACK":
		return table.NewTracks(f.currentLine, f.Score)
	case "SCALE":
		return table.NewScales(f.currentLine, f.Score)
	case "TUNING":
		return table.NewTunings(f.currentLine, f.Score)
	case "TIMBRE":
		return table.NewTimbres(f.currentLine, f.Score)
	case "FILTER":
		return table.NewFilters(f.currentLine, f.Score)
	case "PROPERTY":
		return table.NewProperties(f.currentLine, f.Score)
	default:
		switch name[0] {
		case '=':
			return table.NewSketch(name, f.currentLine, f.Score)
		case '@':
			return table.NewLyrics(name, f.currentLine, f.Score)
		default:
			return table.NewTokens(name, f.currentLine, f.Score)
		}
	}
	//case strings.Contains(rest, "$"):
	//	return &CommandSketch{lineNo: f.currentLine, name: name, score: f.Score, data: rest}
	//default:
	//panic(fmt.Sprintf("invalid line %v: %q in file %q", f.currentLine, line, filepath.Join(f.dir, f.name)))
	//return &ShortCut{lineNo: f.currentLine, name: name, score: f.Score, value: rest}
	//}

	return nil
}

func (f *File) Close() {
	if f.file != nil {
		f.file.Close()
		f.file = nil
	}
}

func (f *File) Parse() (err error) {
	defer f.Close()

	for f.Input.Scan() {
		err = f.Input.Err()
		if err != nil {
			break
		}

		line := strings.TrimSpace(f.Input.Text())

		if len(line) > 1 && line[:2] == "*/" && f.currentMultiLineComment != nil {
			//	fmt.Printf("[%v] multicomment end reached\n", f.currentLine)
			f.currentMultiLineComment = nil
			f.currentLine++
			continue
		}

		part := f.findPart(line)

		if part == nil {
			err = fmt.Errorf("invalid / unexpected line: %q", line)
			//fmt.Printf("got err: %v\n", err)
			break
		}

		//fmt.Printf("[%v] found: %T\n", f.currentLine, part)

		switch v := part.(type) {
		case EmptyLine:
			if f.currentTable != nil {
				err = f.currentTable.Finish()
				if err != nil {
					err = fmt.Errorf("ERROR while finishing table %q: %v", f.currentTable.Name(), err.Error())
					return err
				}
				f.currentTable = nil
			}
			f.Parts = append(f.Parts, v)
			err = v.ParseLine(line)
		case *MultiLineComment:
			if v != f.currentMultiLineComment {
				// fmt.Printf("appending part: %#v\n", v)
				//f.Parts = append(f.Parts, v)
				f.currentMultiLineComment = v
				f.multiComments[f.currentLine] = v
			} else {
				f.currentMultiLineComment.lines = append(f.currentMultiLineComment.lines, line)
			}
		case string:
			f.commentLines[f.currentLine] = v
		case tabler:
			if v != f.currentTable {
				//fmt.Printf("appending part: %#v\n", v)
				f.Parts = append(f.Parts, v)
				f.currentTable = v
			}
			err = v.ParseLine(line)
		case Part:
			//fmt.Printf("appending part: %#v\n", v)
			f.Parts = append(f.Parts, v)
			err = v.ParseLine(line)
		}

		if err != nil {
			//fmt.Printf("ERROR: %v, breaking\n", err)
			//panic(err.Error())
			break
		}

		f.currentLine++
	}

	if err == io.EOF {
		err = nil
	}

	if err != nil {
		return fmt.Errorf("ERROR in line %v: %s", f.currentLine, err.Error())
	}

	if f.currentTable != nil {
		err = f.currentTable.Finish()
		if err != nil {
			err = fmt.Errorf("ERROR while finishing table %q: %v", f.currentTable.Name(), err.Error())
			return err
		}
		f.currentTable = nil
	}

	if f.currentMultiLineComment != nil {
		err = fmt.Errorf("ERROR comment starting at line %v nevers ends", f.currentMultiLineComment.lineNo+1)
	}

	return err
}
