package score

import (
	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/path"
)

var FindInclude func(fsys fs.FS, relDir path.Relative, file string) (resolved path.Relative, err error) = nil

var reservedNames = map[string]bool{
	"MN": true,
	"CC": true,
	"PB": true,
	"AT": true,
	"PT": true,

	"_MN": true,
	"_CC": true,
	"_PB": true,
	"_AT": true,
	"_PT": true,

	"MN_": true,
	"CC_": true,
	"PB_": true,
	"AT_": true,
	"PT_": true,
}
