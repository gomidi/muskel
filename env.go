package muskel

import (
	"fmt"
	"os"
	"path/filepath"

	appdir "github.com/emersion/go-appdir"
	"gitlab.com/gomidi/muskel/predefined"
	"gitlab.com/gomidi/muskel/score"
)

var USER_DIR string
var WORKING_DIR string

func setUserDir(version string) {
	USER_DIR = filepath.Join(appdir.New("muskel").UserConfig(), version)
}

func init() {

	vs := fmt.Sprintf("v%v_%v_%v", VERSION.Major, VERSION.Minor, VERSION.Patch)
	setUserDir(vs)
	setWorkingDir()
	os.MkdirAll(USER_DIR, 0755)
	writePredefinedTemplates()
}

func writePredefinedTemplates() {
	//fmt.Printf("USER_DIR: %q\n", USER_DIR)
	for name, templ := range predefined.SketchesAndTokens {
		p := filepath.Join(USER_DIR, name+".mskl")
		if !score.FileExists(p) {
			os.WriteFile(p, []byte(templ), 0644)
		}
		p = filepath.Join(USER_DIR, name+".md")
		if !score.FileExists(p) {
			os.WriteFile(p, []byte(templ), 0644)
		}
	}

}
