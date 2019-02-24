// +build windows

// set USER_DIR, GLOBAL_DIRS and WORKING_DIR based on the environment variables
// see http://ss64.com/nt/syntax-variables.html
package muskel

import (
	"os"
	"path/filepath"
)

func setUserDir() {
	user_app_data := filepath.ToSlash(os.Getenv("LOCALAPPDATA"))
	if user_app_data == "" {
		user_app_data = filepath.Join(filepath.ToSlash(os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH")), "AppData", "Local", "muskel")
	}
	USER_DIR = filepath.ToSlash(user_app_data)
}

func setWorkingDir() {
	wd, err := os.Getwd()
	if err != nil {
		wd = os.Getenv("CD")
	}

	WORKING_DIR = filepath.ToSlash(wd)
}
