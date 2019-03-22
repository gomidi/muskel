package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type Parser struct {
	Score *Score

	header *HeaderParser
	body   *BodyParser

	input *bufio.Scanner

	current struct {
		isComment bool // track, if we are currently inside a comment
		isHeader  bool // track, if we are currently inside a header
		line      int
	}
}

func NewParser(r io.Reader) *Parser {
	sc := NewScore()
	return &Parser{
		Score: sc,
		input:  bufio.NewScanner(r),
		header: &HeaderParser{sc},
		body: &BodyParser{
			Score:          sc,
			currentBarNo:   -1,
			numInstruments: -1,
		},
	}
}

// ParseFile parses the given file and returns the resulting Score
func ParseFile(filepath string) (sc *Score, err error) {
	b, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	return Parse(bytes.NewReader(b), filepath)
}

var DEBUG bool

// Parse reads from the given reader and returns the resulting Score
func Parse(rd io.Reader, filepath string) (sc *Score, err error) {
	defer func() {
		if !DEBUG {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	ps := NewParser(rd)
	ps.Score.FileName = filepath
	err = ps.Parse()

	if err != nil {
		return nil, err
	}

	return ps.Score, nil
}

// addComment adds a comment to the header or body, depending on whether we are currently inside the header
// or the body
func (p *Parser) addComment(line string) {
	if p.current.isHeader {
		p.header.addComment(line)
		return
	}
	p.body.addComment(p.current.line-p.Score.lineWhereBodyStarts, line)
}

// parseLine parses a line
func (p *Parser) parseLine(header *strings.Builder, line string) error {

	if len(line) > 0 && strings.HasPrefix(line, "=") {
		p.current.isHeader = false
		p.Score.lineWhereBodyStarts = p.current.line + 1
		err := p.header.parseHeader(header.String())
		if err != nil {
			return err
		}
		return nil
	}

	if len(line) == 1 {
		if p.current.isComment {
			p.addComment(line)
			return nil
		}

		if !p.current.isHeader {
			err := p.body.parseLine(string(line))
			if err != nil {
				return fmt.Errorf("Error in line %v:\n%s\n", p.current.line, err.Error())
			}
		}
		return nil
	}

	var start string

	if len(line) > 1 {
		start = string(line[0:2])
	}

	// fmt.Printf("%q %v\n", line, len(line))

	switch start {
	case "//":
		p.addComment(line)
	case "/*":
		p.current.isComment = true
		p.addComment(line)
	case "*/":
		p.current.isComment = false
		p.addComment(line)
	default:
		if p.current.isComment {
			p.addComment(line)
			return nil
		}

		if p.current.isHeader {
			header.WriteString(line + "\n")
			return nil
		}

		err := p.body.parseLine(line)
		if err != nil {
			return fmt.Errorf("Error in line %v:\n%s\n", p.current.line+1, err.Error())
		}
	}
	return nil
}

func (p *Parser) Parse() (err error) {
	var header strings.Builder

	p.current.isHeader = true
	p.current.isComment = false
	p.current.line = -1
	p.Score.lineWhereBodyStarts = 0
	p.body.jumpInLineBefore = true

	for p.input.Scan() {
		p.current.line++

		err = p.input.Err()
		if err != nil {
			break
		}
		p.Score.lastLine++

		err = p.parseLine(&header, p.input.Text())

		if err != nil {
			break
		}
	}

	p.body.finishPart(p.body.currentBarNo + 1)

	if err == io.EOF {
		return nil
	}

	if err != nil {
		fmt.Printf("ERROR while parsing: %v\n", err)
	}

	return err
}
