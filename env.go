package muskel

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

var USER_DIR string
var WORKING_DIR string

func init() {
	setUserDir()
	setWorkingDir()
	os.MkdirAll(USER_DIR, 0755)
	mkdrumnotes()
}

const drumnotes = `
// some shortcuts for drums (GM)
kd: MN36=:
sn: MN40=:
ho: MN46=:
hc: MN42=:
rd: MN51=:
tb: MN54=:
tl: MN45=:
tm: MN48=:
th: MN50=:
sh: MN49=:

=
`

func mkdrumnotes() error {
	p := filepath.Join(USER_DIR, "drumnotes.mskl")
	if !fileExists(p) {
		return ioutil.WriteFile(p, []byte(drumnotes), 0644)
	}
	return nil
}
