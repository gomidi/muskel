package file

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/gomidi/muskel/table"
)

type MultiLineComment struct {
	lineNo int
	lines  []string
}

func (m *MultiLineComment) Start() int {
	return m.lineNo
}

func (m *MultiLineComment) End() int {
	return m.lineNo + len(m.lines) + 1
}

func (m *MultiLineComment) Contains(lineno int) bool {
	if lineno < m.Start() || lineno > m.End() {
		return false
	}
	return true
}

/*
type Property struct {
	lineNo int
	key    string
	value  string
	score  table.Score
}

func (c *Property) LineNo() int {
	return c.lineNo
}

func (c *Property) ParseLine(line string) error {
	pc := strings.SplitN(line, ":", 2)
	c.key = strings.TrimLeft(pc[0], "@")
	c.value = strings.TrimSpace(pc[1])
	c.score.AddProperty(c.key, c.value)
	return nil
}

func (c *Property) WriteTo(f table.Formatter) error {
	return f.WriteLine(fmt.Sprintf("@%s: %s", c.key, c.value))
}
*/

type EmptyLine int

func (e EmptyLine) LineNo() int {
	return int(e)
}

func (e EmptyLine) ParseLine(string) error {
	return nil
}

func (e EmptyLine) WriteTo(f table.Formatter) error {
	return f.WriteLine("")
}

var includeRegExp = regexp.MustCompile("^" + regexp.QuoteMeta("$") + "include")

type Include struct {
	lineNo int
	include
	score
}

type include struct {
	file string
	part string
}

func (i *Include) LineNo() int {
	return i.lineNo
}

func (i *Include) ParseLine(line string) error {
	idx := strings.Index(line, "(")
	if idx < 3 {
		return fmt.Errorf("invalid include line: %q", line)
	}

	args := strings.Split(strings.TrimSpace(strings.TrimRight(line[idx+1:], ")")), ",")

	switch len(args) {
	case 1:
		i.include.file = strings.Trim(strings.TrimSpace(args[0]), `"`)
		i.score.AddInclude(i.include.file, i.include.part, nil)
		return i.score.Include(i.include.file, i.include.part, nil)
	case 2:
		i.include.part = strings.Trim(strings.TrimSpace(args[0]), `"`)
		i.include.file = strings.Trim(strings.TrimSpace(args[1]), `"`)
		i.score.AddInclude(i.include.file, i.include.part, nil)
		return i.score.Include(i.include.file, i.include.part, nil)
	default:
		//return fmt.Errorf("invalid include line: %q", line)
		var params []string
		i.include.part = strings.Trim(strings.TrimSpace(args[0]), `"`)
		i.include.file = strings.Trim(strings.TrimSpace(args[1]), `"`)
		for i := 2; i < len(args); i++ {
			params = append(params, strings.Trim(strings.TrimSpace(args[i]), `"`))
		}
		i.score.AddInclude(i.include.file, i.include.part, params)
		return i.score.Include(i.include.file, i.include.part, params)
	}
}

func (i *Include) WriteTo(f table.Formatter) error {
	var line string
	switch i.include.part {
	case "":
		line = fmt.Sprintf("$include(%q)", i.include.file)
	default:
		line = fmt.Sprintf("$include(%q,%q)", i.include.part, i.include.file)
	}
	return f.WriteLine(line)
}

var embedRegExp = regexp.MustCompile("^" + regexp.QuoteMeta("$") + "embed")

type Embed struct {
	lineNo int
	embed
	score
}

type embed struct {
	params []string
}

func (i *Embed) LineNo() int {
	return i.lineNo
}

func (i *Embed) ParseLine(line string) error {
	idx := strings.Index(line, "(")
	if idx < 3 {
		return fmt.Errorf("invalid include line: %q", line)
	}

	params := strings.Split(strings.TrimSpace(strings.TrimRight(line[idx+1:], ")")), ",")

	i.params = nil

	for _, p := range params {
		i.params = append(i.params, strings.Trim(p, "\""))
	}

	return i.score.Embed(i.params...)
}

func (i *Embed) WriteTo(f table.Formatter) error {
	var line strings.Builder
	line.WriteString("$embed(")

	for n, p := range i.params {
		if n > 0 {
			line.WriteString(",")
		}
		line.WriteString(fmt.Sprintf("%q", p))
	}

	line.WriteString(")")
	return f.WriteLine(line.String())
}

var _ Part = &Include{}
var _ Part = &Embed{}
var _ Part = &table.Table{}
var _ Part = EmptyLine(4)

//var _ Part = &ShortCut{}
//var _ Part = &CommandSketch{}

/*
// ShortCut is a shortcut similar to a sketch column
type ShortCut struct {
	lineNo int
	name   string
	value  string
	score  table.Score
}

func (s *ShortCut) LineNo() int {
	return s.lineNo
}

func (s *ShortCut) ParseLine(line string) error {
	//fmt.Printf("shortcut value: %q\n", s.value)
	//s.value = line
	s.score.AddToken(s.name, s.value)
	return nil
}

func (s *ShortCut) WriteTo(f table.Formatter) error {
	return f.WriteLine(fmt.Sprintf("%s: %s", s.name, s.value))
}
*/

/*
// CommandTable is a command that constructs a table
type CommandSketch struct {
	lineNo int
	name   string
	data   string
	score  table.Score
	cmd    string
	args   []string
}

func (c *CommandSketch) LineNo() int {
	return c.lineNo
}

func (c *CommandSketch) ParseLine(line string) error {
	//fmt.Printf("CommandSketch data: %q\n", c.data)
	line = c.data
	if len(line) < 2 || line[0] != '$' {
		return fmt.Errorf("invalid command: %q", line)
	}

	idx := strings.Index(line, "(")
	if idx < 3 {
		return fmt.Errorf("invalid command: %q", line)
	}

	c.cmd = line[1:idx]
	c.args = strings.Split(strings.TrimSpace(strings.TrimRight(line[idx+1:], ")")), ",")

	tcmd, has := tableCommands[c.cmd]

	if !has {
		return fmt.Errorf("unknown command: %q", c.cmd)
	}

	sk := table.NewSketch(c.name, c.lineNo, c.score)
	return tcmd(sk, c.args...)
}

func (c *CommandSketch) WriteTo(f table.Formatter) error {
	return f.WriteLine(fmt.Sprintf("%s: $%s(%s)", c.name, c.cmd, strings.Join(c.args, ",")))
}

type tableCommand func(sk *table.Sketch, args ...string) error

func euclid(sk *table.Sketch, args ...string) error {
	var eu items.EuclideanRhythm
	err := eu.Parse(0, args...)
	if err != nil {
		return fmt.Errorf("invalid euclidean rhythm: %s", err.Error())
	}

	// add the header
	sk.ParseLine(fmt.Sprintf("%s: | x |", sk.Name()))

	lastPos := (eu.LastPos / 8) + 1

	//fmt.Printf(fmt.Sprintf("#%v/4\n", lastPos))
	sk.ParseLine(fmt.Sprintf("#%v/4", lastPos))

	for _, line := range eu.Sketch {
		sk.ParseLine(fmt.Sprintf("%s | %s |", line[0], line[1]))
	}

	//fmt.Println(eu.PatternDef)

	//lines := strings.Split() eu.PatternDef
	return nil
}

//func(args ...string) *table.Table

var tableCommands = map[string]tableCommand{
	"euclid": euclid,
	// TODO: fill
}
*/
