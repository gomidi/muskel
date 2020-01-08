package score

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/gomidi/muskel/predefined"
)

var USER_DIR string
var WORKING_DIR string

func init() {
	setUserDir()
	setWorkingDir()
	os.MkdirAll(USER_DIR, 0755)
	writePredefinedTemplates()
}

func writePredefinedTemplates() {
	for name, templ := range predefined.SketchesAndTokens {
		p := filepath.Join(USER_DIR, name+".mskl")
		if !FileExists(p) {
			ioutil.WriteFile(p, []byte(templ), 0644)
		}
	}

}
