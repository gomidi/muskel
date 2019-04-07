package template

import (
	"fmt"
	"strconv"
	"strings"
)

type Definition struct {
	Name          string
	NumParams     int
	Original      string
	TimeSignature [2]uint8
}

func (p *Definition) Parse(definitionLine string) error {
	idx := strings.Index(definitionLine, ":")
	if idx < 0 {
		return fmt.Errorf("ERROR in template definition line: missing : to separate name from definition")
	}

	name := strings.TrimSpace(definitionLine[:idx])

	idxAt := strings.Index(name, "@")

	switch {
	case idxAt < 0:
	// do nothing
	case idxAt < 2:
		return fmt.Errorf("ERROR in template definition line: name is too short")
	default:
		tsig := strings.Split(name[idxAt+1:], "/")

		num, errNuM := strconv.Atoi(tsig[0])
		denom, errDenom := strconv.Atoi(tsig[1])

		if errNuM != nil || errDenom != nil || num < 1 || denom < 1 {
			return fmt.Errorf("ERROR in template definition line: invalid time signature")
		}

		p.TimeSignature = [2]uint8{uint8(num), uint8(denom)}
		name = name[:idxAt]
	}

	p.Name = name
	definitionLine = definitionLine[idx+1:]

	var paramsConsolidated = map[string]bool{}

	params := templateReg.FindAllString(definitionLine, -1)
	for _, param := range params {
		paramsConsolidated[param] = true
	}

	p.NumParams = len(paramsConsolidated)
	p.Original = strings.TrimSpace(definitionLine)
	return nil
}

func (p *Definition) Call(params ...string) (string, error) {
	if len(params) != p.NumParams {
		return "", fmt.Errorf("ERROR when calling template definition %q: wrong number of parameters (expected %v, got %v)", p.Name, p.NumParams, len(params))
	}

	if p.NumParams == 0 {
		return p.Original, nil
	}

	// replace params
	s := strings.Replace(p.Original, "%", "%%", -1)
	s = templateReg.ReplaceAllString(s, "%[$1]s")
	ps := make([]interface{}, len(params))

	for i, param := range params {
		ps[i] = param
	}
	return fmt.Sprintf(s, ps...), nil
}
