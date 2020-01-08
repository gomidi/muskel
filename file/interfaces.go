package file

import "gitlab.com/gomidi/muskel/table"

type tabler interface {
	Part
	Name() string
	Cols() []string
	Finish() error
}

type Part interface {
	LineNo() int
	WriteTo(table.Formatter) error
	ParseLine(string) error
}

type score interface {
	table.Score
	Include(file string, sketch string, params []string) error
	HasTrack(name string) bool
	Properties() map[string]interface{}
}
