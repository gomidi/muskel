package table

import (
	"regexp"
)

var DEBUG bool

var _tableNameRegEx = "^(" +
	"[=" + regexp.QuoteMeta(".") + "@]?" +
	"[a-zA-Z]{1}" +
	"[-!~._a-zA-Z0-9]*" +
	")$"

var tableNameRegEx = regexp.MustCompile(_tableNameRegEx)
