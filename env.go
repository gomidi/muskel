package muskel

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/gomidi/muskel/predefined"
	"gitlab.com/gomidi/muskel/score"
)

var USER_DIR string
var WORKING_DIR string

func init() {

	v, err := ParseVersion(VERSION)
	if err != nil {
		panic("invalid VERSION: " + VERSION)
	}
	vs := fmt.Sprintf("v%v_%v_%v", v.Major, v.Minor, v.Patch)
	setUserDir(vs)
	setWorkingDir()
	os.MkdirAll(USER_DIR, 0755)
	writePredefinedTemplates()
}

func writePredefinedTemplates() {
	fmt.Printf("USER_DIR: %q\n", USER_DIR)
	for name, templ := range predefined.SketchesAndTokens {
		p := filepath.Join(USER_DIR, name+".mskl")
		if !score.FileExists(p) {
			ioutil.WriteFile(p, []byte(templ), 0644)
		}
	}

}
