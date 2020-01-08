package sketch

import "regexp"

var DEBUG bool = false
var TemplateReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([0-9]+)")

var disallowedColNames = map[string]bool{
	".": true,
	"~": true,
	"-": true,
	"*": true,
	"+": true,
}
