package muskel

import (
	"fmt"
	"strings"
)

type TemplateDefinition struct {
	Name      string
	NumParams int
	Original  string
}

func (p *TemplateDefinition) Parse(definitionLine string) error {
	idx := strings.Index(definitionLine, ":")
	if idx < 0 {
		return fmt.Errorf("ERROR in template definition line: missing : to separate name from definition")
	}

	p.Name = strings.TrimSpace(definitionLine[:idx])
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

func (p *TemplateDefinition) Call(params ...string) (string, error) {
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
