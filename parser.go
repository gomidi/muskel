package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

type Parser struct {
	Score *Score

	header *HeaderParser
	body   *BodyParser

	input *bufio.Reader

	current struct {
		isComment bool // track, if we are currently inside a comment
		isHeader  bool // track, if we are currently inside a header
		line      int
	}

	//	lineWhereBodyStarts int
}

func NewParser(r io.Reader) *Parser {
	sc := NewScore()
	return &Parser{
		Score:  sc,
		input:  bufio.NewReader(r),
		header: &HeaderParser{sc},
		body: &BodyParser{
			Score:          sc,
			currentBar:     -1,
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

	return Parse(bytes.NewReader(b))
}

// Parse reads from the given reader and returns the resulting Score
func Parse(rd io.Reader) (sc *Score, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	ps := NewParser(rd)
	err = ps.Parse()

	if err != nil {
		return nil, err
	}

	return ps.Score, nil
}

// addComment adds a comment to the header or body, depending on whether we are currently inside the header
// or the body
func (p *Parser) addComment(line []byte) {
	if p.current.isHeader {
		p.header.addComment(line)
		return
	}
	p.body.addComment(p.current.line-p.Score.lineWhereBodyStarts, line)
}

// parseLine parses a line
func (p *Parser) parseLine(header *bytes.Buffer, line []byte) error {

	if len(line) > 0 && string(line[0]) == "=" {
		p.current.isHeader = false
		p.Score.lineWhereBodyStarts = p.current.line + 1
		return p.header.parseHeader(header.String())
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

	switch string(line[0:2]) {
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
			header.Write(line)
			header.WriteString("\n")
			return nil
		}
		err := p.body.parseLine(string(line))
		if err != nil {
			return fmt.Errorf("Error in line %v:\n%s\n", p.current.line+1, err.Error())
		}
	}
	return nil
}

func (p *Parser) readLine(header *bytes.Buffer) error {
	p.current.line++
	line, _, err := p.input.ReadLine()

	if err != nil {
		return err
	}
	p.Score.lastLine++

	return p.parseLine(header, line)
}

func (p *Parser) Parse() (err error) {
	var header = &bytes.Buffer{}

	p.current.isHeader = true
	p.current.isComment = false
	p.current.line = -1
	p.Score.lineWhereBodyStarts = 0

	for {
		err = p.readLine(header)

		if err != nil {
			break
		}
	}

	if err == io.EOF {
		return nil
	}

	if err != nil {
		fmt.Printf("ERROR while parsing: %v\n", err)
	}

	return err
}
