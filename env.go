package muskel

import (
	"fmt"
	"os"
	"path/filepath"

	appdir "github.com/emersion/go-appdir"
	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"
	"gitlab.com/gomidi/muskel/predefined"
)

var USER_DIR string
var WORKING_DIR string

func setUserDir(version string) {
	USER_DIR = filepath.Join(appdir.New("muskel").UserConfig(), version)
}

func init() {
	v := version.BuildVersion()
	vs := "v0_0_0"
	if v != nil {
		vs = fmt.Sprintf("v%v_%v_%v", v.Major, v.Minor, v.Patch)
	}
	setUserDir(vs)
	setWorkingDir()
	os.MkdirAll(USER_DIR, 0755)
	err := writePredefinedTemplates()

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't initialize predefined templates: %s\n", err.Error())
	}
}

func writePredefinedTemplates() error {
	fsys, err := rootfs.New()

	if err != nil {
		return err
	}

	return WritePredefinedTemplates(fsys)
}

func WritePredefinedTemplates(fsys fs.FS) error {
	dir := path.MustDirFromSystem(USER_DIR).RootRelative()

	if !path.IsDir(dir) {
		return fs.ErrExpectedDir.Params(dir.String())
	}

	if !fsys.Exists(dir) {
		err := fs.MkDirAll(fsys, dir)
		if err != nil {
			return err
		}
		//return fs.ErrNotFound.Params(dir.String())
	}

	//fmt.Printf("USER_DIR: %q\n", USER_DIR)
	for name, templ := range predefined.SketchesAndTokens {

		p := dir.Join(name + ".mskl")
		if !fsys.Exists(p) {
			err := fs.WriteFile(fsys, p, []byte(templ), false)
			if err != nil {
				return err
			}
		}

		p = dir.Join(name + ".md")
		if !fsys.Exists(p) {
			err := fs.WriteFile(fsys, p, []byte(templ), false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
