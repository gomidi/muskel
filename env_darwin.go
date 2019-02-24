// +build darwin

package muskel

/*
according to http://wiki.freepascal.org/Multiplatform_Programming_Guide#Configuration_files
/etc
/Users/user/.config/project1
*/

import (
	"os"
	"path/filepath"
)

func setUserDir() {
	home := os.Getenv("HOME")
	if home == "" {
		home = filepath.Join("/home", os.Getenv("USER"))
	}
	USER_DIR = filepath.Join(home + ".muskel")
}

func setWorkingDir() {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	WORKING_DIR = wd
}
