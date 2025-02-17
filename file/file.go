package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.com/gomidi/muskel/table"
)

type File struct {
	Score score
	Input *bufio.Scanner
	Parts []Part

	name                    string
	dir                     string
	file                    []byte
	multiComments           map[int]*MultiLineComment
	commentLines            map[int]string
	currentLine             int
	currentTable            tabler
	currentMultiLineComment *MultiLineComment
	output                  io.Writer
}

func (f *File) SetFileInfos(file []byte, fpath string) {
	f.file = file
	f.dir = filepath.Dir(fpath)
	f.name = filepath.Base(fpath)
}

func (f *File) Name() string {
	return filepath.Join(f.dir, f.name)
}

func (f *File) writeComment() (bool, error) {
	if cmt, has := f.commentLines[f.currentLine]; has {
		err := f.println(cmt, true)
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
				err := f.println("/*", true)
				return true, err
			}

			if f.currentLine == mcomment.End() {
				err := f.println("*/", true)
				return false, err
			}
			err := f.println(mcomment.lines[f.currentLine-start-1], true)
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
	return f.println(line, true)
}

func (f *File) WriteDashLine(line string) (err error) {
	return f.println(line, false)
}

func (f *File) println(line string, count bool) error {
	_, err := fmt.Fprintln(f.output, line)
	//fmt.Fprintf(os.Stderr, "printing line %v: %q\n", f.currentLine, line)
	if count {
		f.currentLine++
	}
	return err
}

func (f *File) Format() error {
	fpath := filepath.Join(f.dir, f.name)
	os.Remove(fpath)

	fl, err := os.Create(fpath)
	if err != nil {
		return err
	}
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

	sketches := map[string]*table.Sketch{}
	tokens := map[string]*table.Tokens{}
	var parts []Part

	for _, part := range f.Parts {
		//fmt.Printf("parttype: %T\n", part)
		switch v := part.(type) {
		case *table.Sketch:
			if sk, has := sketches[v.Name()]; has {
				sk.Merge(v)
			} else {
				sketches[v.Name()] = v
				parts = append(parts, part)
			}
		case *table.Tokens:
			tokens[v.Name()] = v
			parts = append(parts, part)
		default:
			parts = append(parts, part)
		}
	}

	for _, part := range parts {
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

var dashLine = regexp.MustCompile(`^\s*\|\s*-+\s*\|`)

func (f *File) findPart(line string) (interface{}, error) {
	if len(line) > 1 && line[:2] == "/*" {
		if f.currentMultiLineComment != nil {
			return nil, nil
		}
		return &MultiLineComment{lineNo: f.currentLine}, nil
	}

	if f.currentMultiLineComment != nil {
		return f.currentMultiLineComment, nil
	}

	/*
		if len(line) > 1 && line[0] == '@' {
			return &Property{lineNo: f.currentLine, score: f.Score}
		}
	*/

	if len(line) > 1 && line[:2] == "//" {
		return line, nil
	}

	if strings.TrimSpace(line) == "" {
		return EmptyLine(f.currentLine), nil
	}

	if f.currentTable != nil {
		//hmm := strings.ReplaceAll(strings.ReplaceAll(line, " ",""), "|", "")
		if dashLine.MatchString(line) {
			f.currentLine--
		}

		return f.currentTable, nil
	}

	if includeRegExp.MatchString(strings.ToLower(line)) {
		return &Include{
			lineNo: f.currentLine,
			score:  f.Score,
		}, nil
	}

	if embedRegExp.MatchString(strings.ToLower(line)) {
		return &Embed{
			lineNo: f.currentLine,
			score:  f.Score,
		}, nil
	}

	//fmt.Printf("line: %q\n", line)

	trimmedLine := strings.TrimSpace(line)

	idx := strings.Index(trimmedLine, "|")

	if idx == 0 {

		name := strings.TrimSpace(trimmedLine[1:])

		//name := strings.TrimSpace(line)
		nextIdx := strings.Index(name, "|")
		if nextIdx > 1 {
			name = strings.TrimSpace(name[:nextIdx])
		}

		//fmt.Printf("name: %q\n", name)
		//rest := strings.TrimSpace(line[idx+1:])

		//switch {
		//case strings.Contains(rest, "|"):

		currentLine := f.currentLine
		//f.currentLine--

		switch name {
		case "TRACK":
			return table.NewTracks(currentLine, f.Score), nil
		case "SCALE":
			return table.NewScales(currentLine, f.Score), nil
		case "TUNING":
			return table.NewTunings(currentLine, f.Score), nil
		case "TIMBRE":
			return table.NewTimbres(currentLine, f.Score), nil
		case "FILTER":
			return table.NewFilters(currentLine, f.Score), nil
		case "PROPERTY":
			return table.NewProperties(currentLine, f.Score), nil
		default:
			switch name[0] {
			case '=':
				return table.NewSketch(name, currentLine, f.Score), nil
			case '@':
				return table.NewLyrics(name, currentLine, f.Score), nil
			case '.':
				return table.NewTokens(name, currentLine, f.Score), nil
			default:
				return nil, fmt.Errorf("invalid table name: %q", name)
			}
		}

	} else {
		// just another possbly markdown syntax line
		return line, nil
	}
	//case strings.Contains(rest, "$"):
	//	return &CommandSketch{lineNo: f.currentLine, name: name, score: f.Score, data: rest}
	//default:
	//panic(fmt.Sprintf("invalid line %v: %q in file %q", f.currentLine, line, filepath.Join(f.dir, f.name)))
	//return &ShortCut{lineNo: f.currentLine, name: name, score: f.Score, value: rest}
	//}

	//return nil, nil
}

type ParseError struct {
	File string
	Line int
	Err  error
}

func (p ParseError) Error() string {
	return fmt.Sprintf("%s:%v: %v", filepath.FromSlash(filepath.ToSlash(p.File)), p.Line, p.Err)
}

func (f *File) Parse() (err error) {
	var perr ParseError
	//perr.File = f.Name()
	perr.File = f.name

	for f.Input.Scan() {
		perr.Line = f.currentLine

		perr.Err = f.Input.Err()
		if perr.Err != nil {
			//fmt.Printf("error while scanning: %v\n", err)
			break
		}

		line := strings.TrimSpace(f.Input.Text())
		//fmt.Printf("[%v] line: %q\n", f.currentLine, line)

		if len(line) > 1 && line[:2] == "*/" && f.currentMultiLineComment != nil {
			//	fmt.Printf("[%v] multicomment end reached\n", f.currentLine)
			f.currentMultiLineComment = nil
			f.currentLine++
			continue
		}

		var part interface{}

		part, perr.Err = f.findPart(line)

		if perr.Err != nil {
			perr.Line = perr.Line + 1
			break
		}

		if part == nil {
			perr.Err = fmt.Errorf("invalid / unexpected line: %q", line)
			break
		}

		//fmt.Printf("[%v] found: %T\n", f.currentLine, part)

		switch v := part.(type) {
		case EmptyLine:
			if f.currentTable != nil {
				perr.Err = f.currentTable.Finish()
				if perr.Err != nil {
					perr.Err = fmt.Errorf("ERROR while finishing table %q: %v", f.currentTable.Name(), perr.Err.Error())
					return perr
				}
				f.currentTable = nil
			}
			f.Parts = append(f.Parts, v)
			perr.Err = v.ParseLine(line)
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
			line = strings.TrimSpace(line)

			if !dashLine.MatchString(line) {
				if strings.LastIndex(line, "|") != len(line)-1 {
					line += "|"
				}
				perr.Err = v.ParseLine(line)
			}
		case Part:
			//fmt.Printf("appending part: %#v\n", v)
			f.Parts = append(f.Parts, v)
			line = strings.TrimSpace(line)

			if !dashLine.MatchString(line) {
				if strings.LastIndex(line, "|") != len(line)-1 && len(line) > 0 && line[0] != '\'' {
					line += "|"
				}
				perr.Err = v.ParseLine(line)
			}
		}

		if perr.Err != nil {
			//panic(err.Error())
			break
		}

		f.currentLine++
	}

	if perr.Err == io.EOF {
		perr.Err = nil
	}

	if perr.Err != nil {
		return perr
	}

	if f.currentTable != nil {
		perr.Err = f.currentTable.Finish()
		if perr.Err != nil {
			perr.Err = fmt.Errorf("ERROR while finishing table %q: %v", f.currentTable.Name(), perr.Err.Error())
			return perr
		}
		f.currentTable = nil
	}

	if f.currentMultiLineComment != nil {
		perr.Err = fmt.Errorf("ERROR comment starting at line %v nevers ends", f.currentMultiLineComment.lineNo+1)
	}

	if perr.Err == nil {
		return nil
	}

	return perr
}
