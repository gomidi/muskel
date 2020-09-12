package score

var FindInclude func(relDir string, file string) (resolved string, err error) = nil

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
