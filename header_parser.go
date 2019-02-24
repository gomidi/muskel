package muskel

import (
	"fmt"
	"regexp"
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
	err := pd.Parse(line)
	if err != nil {
		return fmt.Errorf("invalid pattern definition line %#v: %s", line, err)
	}
	p.Score.PatternDefinitions[pd.Name] = &pd
	return nil
}

// parseMetaLine parses a line in the header that contains a meta property
func (p *HeaderParser) parseMetaLine(line string) error {
	//	fmt.Printf("parsing meta line %q\n", line)
	keyVal := strings.Split(line, ":")
	if len(keyVal) != 2 {
		return fmt.Errorf("invalid header line %#v, format must be key: value", line)
	}

	key := strings.ToLower(strings.TrimSpace(keyVal[0]))
	val := strings.TrimSpace(keyVal[1])

	p.Score.Meta[key] = val
	//	fmt.Printf("setting meta %q -> %q\n", key, val)
	return nil
}

var regExPattern = regexp.MustCompile("^([a-zA-Z][a-zA-Z]+).*")

func (p *HeaderParser) includeScore(sc *Score) error {
	//panic("implement")
	for k, v := range sc.PatternDefinitions {
		if p.Score.PatternDefinitions[k] != nil || p.Score.IncludedPatternDefinitions[k] != nil {
			return fmt.Errorf("pattern %q already defined in main file, can't redefine", k)
		}
		p.Score.IncludedPatternDefinitions[k] = v
	}
	return nil
}

func (p *HeaderParser) include(file string) error {
	p.Score.HeaderIncludes = append(p.Score.HeaderIncludes, file)
	sc, err := p.Score.include(file)
	if err != nil {
		return fmt.Errorf("Error while including file %q: %s", file, err.Error())
	}

	err = p.includeScore(sc)
	if err != nil {
		return fmt.Errorf("Error while including file %q: %s", file, err.Error())
	}
	return nil
}

func (p *HeaderParser) parseCommand(data string) error {
	var c CommandCall
	err := c.Parse(data)
	if err != nil {
		return err
	}
	switch c.Name {
	case "include":
		if len(c.Params) != 1 {
			return fmt.Errorf("include command needs one parameter")
		}
		return p.include(strings.Trim(c.Params[0], "\""))
	default:
		return fmt.Errorf("unsupported command in body: %q", c.Name)
	}
}

// parseHeaderLine parses a line of the header
func (p *HeaderParser) parseHeaderLine(line string) error {
	if len(line) == 0 {
		return nil
	}

	switch line[0] {
	case '/':
		return p.parseTemperamentLine(line)
	case '@':
		return p.parseMetaLine(line[1:])
	case '$':
		return p.parseCommand(line[1:])
	default:
		if regExPattern.MatchString(line) {
			return p.parsePatternDefinitionLine(line)
		}
		return nil
	}
}

// parseHeader parses the header (everything up to the  =)
func (p *HeaderParser) parseHeader(data string) error {
	//	fmt.Printf("parsing header line: %q\n", data)
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		err := p.parseHeaderLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}
