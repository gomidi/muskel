package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/gomidi/muskel/muskellib"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/template"
)

type header struct {
	*score.Score
}

func (p *header) addComment(line string) {
	p.Score.HeaderComments = append(p.Score.HeaderComments, line)
}

// parseTemperamentLine parses a line of temperament definition (in the header)
func (p *header) parseTemperamentLine(line string) error {
	keyVal := strings.Split(line[1:], "/")
	if len(keyVal) != 2 {
		return fmt.Errorf("invalid header line %#v, format must be /note/frequency", line)
	}

	p.Score.Temperament[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
	return nil
}

// parseTemplateDefinitionLine parses a line in the header that defines a template
func (p *header) parseTemplateDefinitionLine(line string) error {
	var pd template.Definition
	err := pd.Parse(line)
	if err != nil {
		return fmt.Errorf("invalid template definition line %#v: %s", line, err)
	}
	p.Score.TemplateDefinitions[pd.Name] = &pd
	return nil
}

// parseMetaLine parses a line in the header that contains a meta property
func (p *header) parseMetaLine(line string) error {
	//	fmt.Printf("parsing meta line %q\n", line)
	keyVal := strings.SplitN(line, ":", 2)
	if len(keyVal) != 2 {
		return fmt.Errorf("invalid header line %#v, format must be key: value", line)
	}

	key := strings.ToLower(strings.TrimSpace(keyVal[0]))
	val := strings.TrimSpace(keyVal[1])

	p.Score.Meta[key] = val
	//	fmt.Printf("setting meta %q -> %q\n", key, val)
	return nil
}

var regExTemplate = regexp.MustCompile("^([a-zA-Z][_a-zA-Z]+).*")

func (p *header) includeScore(sc *score.Score) error {
	for k, v := range sc.TemplateDefinitions {
		if p.Score.TemplateDefinitions[k] != nil || p.Score.IncludedTemplateDefinitions[k] != nil {
			return fmt.Errorf("template %q already defined in main file, can't redefine", k)
		}
		p.Score.IncludedTemplateDefinitions[k] = v
	}
	return nil
}

func (p *header) include(file string) error {
	p.Score.HeaderIncludes = append(p.Score.HeaderIncludes, file)
	sc, err := IncludeScore(p.Score, file)
	if err != nil {
		return fmt.Errorf("Error while including file %q: %s", file, err.Error())
	}

	err = p.includeScore(sc)
	if err != nil {
		return fmt.Errorf("Error while including file %q: %s", file, err.Error())
	}
	return nil
}

func (p *header) parseCommand(data string) error {
	var c muskellib.CommandCall
	err := c.Parse(strings.TrimSpace(data))
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
func (p *header) parseHeaderLine(line string) error {
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
		if regExTemplate.MatchString(line) {
			return p.parseTemplateDefinitionLine(line)
		}
		return nil
	}
}

// parseHeader parses the header (everything up to the  =)
func (p *header) parseHeader(data string) error {

	sc := bufio.NewScanner(strings.NewReader(data))
	//	fmt.Printf("parsing header line: %q\n", data)
	//lines := strings.Split(data, "\n")

	for sc.Scan() {
		err := sc.Err()

		if err != nil {
			return err
		}

		err = p.parseHeaderLine(sc.Text())

		if err != nil {
			return err
		}
	}

	return nil
}
