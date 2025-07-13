package file

import (
	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel/table"
)

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
	Include(file path.Relative, sketch string, params []string) error
	Embed(patterns ...string) error
	HasTrack(name string) bool
	Properties() map[string]string
	GetFS() fs.FS
}
