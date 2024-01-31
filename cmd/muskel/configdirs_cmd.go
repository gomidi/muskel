package main

import (
	"fmt"
	"os"

	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/gomidi/muskel"
)

type configDirsCmd struct {
	*config.Config
}

var CONFIG_DIRS = &configDirsCmd{}

func init() {
	CONFIG_DIRS.init()
}

func (s *configDirsCmd) init() {
	s.Config = CONFIG.Command("dirs", "show configuration dirs").SkipAllBut()
}

func (c *configDirsCmd) run() error {
	fmt.Fprintf(os.Stdout, "USER_DIR: %q\nWORKING_DIR: %q\n\n", muskel.USER_DIR, muskel.WORKING_DIR)
	return nil
}
