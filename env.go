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
@partial: header

// some shortcuts for drums (GM)
kd: 1MN36
sn: 1MN40
ho: 1MN46
hc: 1MN42
rd: 1MN51
tb: 1MN54
tl: 1MN45
tm: 1MN48
th: 1MN50
sh: 1MN49

=
`

func mkdrumnotes() error {
	p := filepath.Join(USER_DIR, "drumnotes.mskl")
	if !fileExists(p) {
		return ioutil.WriteFile(p, []byte(drumnotes), 0644)
	}
	return nil
}
