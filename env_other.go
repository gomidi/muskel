// +build !linux,!windows,!darwin

package muskel

// environment for unixy system that are not linux and not darwin, like the BSD family

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
