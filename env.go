package muskel

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	appdir "github.com/emersion/go-appdir"
	"gitlab.com/gomidi/muskel/predefined"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/metakeule/config"
)

var USER_DIR string
var WORKING_DIR string

func setUserDir(version string) {
	USER_DIR = filepath.Join(appdir.New("muskel").UserConfig(), version)
}

func init() {

	v, err := config.ParseVersion(VERSION)
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
	//fmt.Printf("USER_DIR: %q\n", USER_DIR)
	for name, templ := range predefined.SketchesAndTokens {
		p := filepath.Join(USER_DIR, name+".mskl")
		if !score.FileExists(p) {
			ioutil.WriteFile(p, []byte(templ), 0644)
		}
		p = filepath.Join(USER_DIR, name+".md")
		if !score.FileExists(p) {
			ioutil.WriteFile(p, []byte(templ), 0644)
		}
	}

}
