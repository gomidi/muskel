package muskel

import (
	"fmt"
	"strings"
)

type HeaderParser struct {
	*Score
}

func (p *HeaderParser) addComment(line []byte) {
	p.Score.HeaderComments = append(p.Score.HeaderComments, string(line))
}

// parseTemperamentLine parses a line of temperament definition (in the header)
func (p *HeaderParser) parseTemperamentLine(line string) error {
	keyVal := strings.Split(line[1:], "/")
	if len(keyVal) != 2 {
		return fmt.Errorf("invalid header line %#v, format must be /note/frequency", line)
	}

	p.Score.Temperament[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
	return nil
}

// parsePatternDefinitionLine parses a line in the header that defines a pattern
func (p *HeaderParser) parsePatternDefinitionLine(line string) error {
	var pd PatternDefinition
	err := pd.Parse(line[1:])
	if err != nil {
		return fmt.Errorf("invalid pattern definition line %#v: %s", line, err)
	}
	p.Score.PatternDefinitions[pd.Name] = &pd
	return nil
}

// parseMetaLine parses a line in the header that contains a meta property
func (p *HeaderParser) parseMetaLine(line string) error {
	keyVal := strings.Split(line, ":")
	if len(keyVal) != 2 {
		return fmt.Errorf("invalid header line %#v, format must be key: value", line)
	}

	key := strings.ToLower(strings.TrimSpace(keyVal[0]))
	val := strings.TrimSpace(keyVal[1])

	p.Score.Meta[key] = val
	return nil
}

// parseHeaderLine parses a line of the header
func (p *HeaderParser) parseHeaderLine(line string) error {
	if len(line) == 0 {
		return nil
	}

	switch line[0] {
	case '/':
		return p.parseTemperamentLine(line)
	case '$':
		return p.parsePatternDefinitionLine(line)
	default:
		return p.parseMetaLine(line)
	}
}

// parseHeader parses the header (everything up to the  =)
func (p *HeaderParser) parseHeader(data string) error {
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		err := p.parseHeaderLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}
