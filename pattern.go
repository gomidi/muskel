package muskel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type PatternDefinition struct {
	Name      string
	NumParams int
	Original  string
}

type PatternCall struct {
	Name         string
	Params       []string
	Replacements []string
	Slice        [2]int
	SyncFirst    bool
	result       string
	Events       [][]*positionedEvent
}

type positionedEvent struct {
	position string
	item     interface{}
}

func (p *PatternCall) parseItem(data string) (item interface{}, err error) {
	return (&itemParser{nil}).parseItem(data)
}

func (p *PatternCall) parseItems(data string) (item interface{}, err error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch data[0] {
	case '(':
		var items MultiItem
		data = strings.Trim(data, "()")
		its := strings.Split(data, ",")

		for _, it := range its {
			it = strings.TrimSpace(it)

			if it != "" {
				ite, err := p.parseItem(it)
				if err != nil {
					return nil, err
				}
				items = append(items, ite)
			}
		}

		return items, nil

	default:
		return p.parseItem(data)
	}

	return
}

func (p *PatternCall) parseEvent(data string) (ev *positionedEvent, err error) {

	ev = &positionedEvent{}
	var positionBf bytes.Buffer
	var bf bytes.Buffer
	rd := bufio.NewReader(strings.NewReader(strings.TrimSpace(data)))
	_ = rd
	var rn rune

	for err == nil {
		rn, _, err = rd.ReadRune()

		if err != nil {
			break
		}

		switch {
		case strings.IndexRune("123456789&;:.", rn) >= 0:
			positionBf.WriteRune(rn)
		default:
			bf.WriteRune(rn)
		}

	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	ev.position = positionBf.String()

	item, err := p.parseItems(bf.String())

	ev.item = item

	if err != nil {
		return nil, err
	}

	return ev, nil
}

func (p *PatternCall) parseBars(data string) error {
	bars := strings.Split(data, "|")
	p.Events = make([][]*positionedEvent, len(bars))

	for bidx, bar := range bars {
		evts := strings.Split(strings.TrimSpace(bar), " ")

		for _, ev := range evts {
			ev = strings.TrimSpace(ev)

			if ev != "" {
				e, err := p.parseEvent(ev)
				if err != nil {
					return fmt.Errorf("could not parse event %q: %s", ev, err.Error())
				}
				p.Events[bidx] = append(p.Events[bidx], e)
			}
		}
	}
	return nil
}

func splitBarsAndItems(def string) (all [][]string) {
	def = strings.TrimSpace(def)
	bars := strings.Split(def, "|")

	for _, b := range bars {
		b = strings.TrimSpace(b)
		its := strings.Split(b, " ")
		var items []string

		for _, it := range its {
			it = strings.TrimSpace(it)
			if it != "" {
				items = append(items, it)
			}
		}

		if len(items) > 0 {
			all = append(all, items)
		}
	}

	return
}

// syntax for params:  %[1] %[2] etc.

//var pattReg = regexp.MustCompile("%" + regexp.QuoteMeta("[") + "([1-9]+)" + regexp.QuoteMeta("]"))
var pattReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([1-9]+)")

/*
type Getter interface {
	GetPatternDefinition(name string) *PatternDefinition
}
*/

var regPos = regexp.MustCompile("^([1-9]?[0-9]?)([&;" + regexp.QuoteMeta(".") + "]*)")

//var regNote = regexp.MustCompile(`^([a-gA-G])([#\\^°]?)(["']*)$`)

type patternFragment struct {
	position string
	item     string
}

func (f *patternFragment) parse(s string) {
	if regPos.MatchString(s) {
		all := regPos.FindAllString(s, 1)
		f.position = all[0]

		s = regPos.ReplaceAllString(s, "")
	}
	f.item = s
}

func replaceItemWith(src string, replacement string) string {
	var fsrc patternFragment
	var frepl patternFragment
	fsrc.parse(src)
	frepl.parse(replacement)

	if frepl.position != "" {
		fsrc.position = frepl.position
	}

	if frepl.item != "" {
		fsrc.item = frepl.item
	}

	return fsrc.position + fsrc.item

	//return replacement
}

//var regPatternOverwrite = regexp.MustCompile(`^([1-9]?[0-9]?)(&?)(;?)(\\.?)(;?)([\\+\\-]*)([a-gA-G]*)(#|\\^|°)?(["']*)`)

/*
var regPatternCall = regexp.MustCompile(
"^([a-zA-Z][0-9a-zA-Z]*)"+
"("+regexp.QuoteMeta("[") + "[1-9]+" + regexp.QuoteMeta("]")+")"+
regexp.QuoteMeta("/") + "[^/]" + regexp.QuoteMeta("/")
)
*/

func (p *PatternDefinition) Parse(definitionLine string) error {
	idx := strings.Index(definitionLine, ":")
	if idx < 0 {
		return fmt.Errorf("ERROR in pattern definition line: missing : to separate name from definition")
	}

	p.Name = strings.TrimSpace(definitionLine[:idx])
	definitionLine = definitionLine[idx+1:]
	var paramsConsolidated = map[string]bool{}
	params := pattReg.FindAllString(definitionLine, -1)
	for _, param := range params {
		paramsConsolidated[param] = true
	}
	p.NumParams = len(paramsConsolidated)
	p.Original = strings.TrimSpace(definitionLine)
	return nil
}

func (p *PatternDefinition) Call(call *PatternCall, getter func(name string) *PatternDefinition) (string, error) {
	s := p.Original

	// replace params
	if p.NumParams > 0 {
		//if len(call.Params) > p.NumParams {
		if len(call.Params) != p.NumParams {
			//return "", fmt.Errorf("ERROR when calling pattern definition %q: too many parameters (expected %v, got %v)", p.Name, p.NumParams,len(call.Params))
			return "", fmt.Errorf("ERROR when calling pattern definition %q: wrong number of parameters (expected %v, got %v)", p.Name, p.NumParams, len(call.Params))
		}
		//		fmt.Printf("placeholder not replaced: %q\n", s)
		s = strings.Replace(s, "%", "%%", -1)
		s = pattReg.ReplaceAllString(s, "%[$1]s")
		//		fmt.Printf("placeholder replaced: %q\n", s)
		params := make([]interface{}, len(call.Params))

		for i, param := range call.Params {
			params[i] = param
		}
		s = fmt.Sprintf(s, params...)
		//		fmt.Printf("after sprintf: %q\n", s)
	}

	all := splitBarsAndItems(s)

	// we must parse each item and if we have a pattern call, we must the pattern we get via the getter
	for bidx, bar := range all {
		for itidx, item := range bar {
			if len(item) > 1 && item[0] == '$' {
				var pc PatternCall
				pc.Parse(item[1:])

				pd := getter(pc.Name)
				if pd == nil {
					return "", fmt.Errorf("ERROR while calling pattern %q: can't find referenced pattern %q", p.Name, pc.Name)
				}

				res, err := pd.Call(&pc, getter)

				if err != nil {
					return "", fmt.Errorf("ERROR while calling pattern %q: can't replace referenced pattern %q: %s", p.Name, pc.Name, err)
				}

				bar[itidx] = res
				all[bidx] = bar
			}
		}
	}

	// slice the items
	if call.Slice[0] > -1 {
		first := call.Slice[0]
		last := call.Slice[1]

		var res [][]string

		i := 0
		collecting := false
		res = append(res, []string{})

		for _, bar := range all {
			for _, item := range bar {
				if i == first {
					collecting = true
				}
				if last > 0 && i >= last {
					collecting = false
				}
				if collecting {
					res[len(res)-1] = append(res[len(res)-1], item)
				}
				i++
			}

			if collecting {
				res = append(res, []string{})
			}
		}

		all = res[:len(res)-1]
	}

	// do the replacement
	lenRepl := len(call.Replacements)

	if lenRepl > 0 {

		/*
			for ridx, repl := range call.Replacements {

			}
		*/

		i := 0

		for bidx, bar := range all {
			for itidx, item := range bar {
				if i < lenRepl && call.Replacements[i] != ":" {
					bar[itidx] = replaceItemWith(item, call.Replacements[i])
				}
				i++
			}
			all[bidx] = bar
		}
	}

	var endRes []string

	for _, bar := range all {
		endRes = append(endRes, strings.Join(bar, " "))
	}

	s = strings.Join(endRes, "|")

	//s = strings.Join()

	return s, nil

	//return ""
}

func (p *PatternCall) Parse(call string) error {
	replacements := ""
	slice := ""
	params := ""

	if idx := strings.Index(call, "/"); idx > 0 {
		replacements = strings.TrimSpace(strings.Trim(call[idx:], "/"))
		call = call[:idx]
	}

	if idx := strings.Index(call, "["); idx > 0 {
		slice = strings.TrimSpace(strings.Trim(call[idx:], "[]"))
		call = call[:idx]
	}

	if idx := strings.Index(call, "("); idx > 0 {
		params = strings.TrimSpace(strings.Trim(call[idx:], "()"))
		call = call[:idx]
	}

	p.Name = call

	if p.Name[0] == '\\' {
		p.SyncFirst = true
		p.Name = p.Name[1:]
	}

	if replacements != "" {
		repl := strings.Split(replacements, " ")

		for _, rep := range repl {
			rep = strings.TrimSpace(rep)

			if rep != "" {
				p.Replacements = append(p.Replacements, rep)
			}
		}
	}

	if params != "" {
		ps := strings.Split(params, ",")

		for _, param := range ps {
			param = strings.TrimSpace(param)

			if param != "" {
				p.Params = append(p.Params, param)
			}
		}
	}

	p.Slice[0] = -1
	p.Slice[1] = -1

	if slice != "" {
		sl := strings.Split(slice, ":")
		if len(sl) != 2 {
			return fmt.Errorf("ERROR in call of pattern %q: invalid slice %q", p.Name, "["+slice+"]")
		}

		from := strings.TrimSpace(sl[0])
		to := strings.TrimSpace(sl[1])

		if from == "" {
			p.Slice[0] = 0
		} else {
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return fmt.Errorf("ERROR in call of pattern %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[0] = fromI
		}

		if to != "" {
			toI, err := strconv.Atoi(to)
			if err != nil || toI == 0 {
				return fmt.Errorf("ERROR in call of pattern %q: invalid slice %q", p.Name, "["+slice+"]")
			}
			p.Slice[1] = toI
		}

	}
	return nil
}
