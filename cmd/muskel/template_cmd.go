package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/metakeule/config"
)

type templateCmd struct {
	*config.Config
	File config.StringGetter
}

var TEMPLATE = templateCmd{}

func init() {
	TEMPLATE.init()
}

func (s *templateCmd) init() {
	s.Config = CONFIG.MustCommand("template", "show global template files").SkipAllBut()
	s.File = s.LastString("templatefile", "file name of the template file that should be shown. If no file is given, the list of available files is shown.")
}

func (t *templateCmd) run() error {
	if t.File.IsSet() && t.File.Get() != "" {
		return t.printFile(filepath.Join(muskel.USER_DIR, t.File.Get()+muskel.FILE_EXTENSION))
	} else {
		return t.printDir(muskel.USER_DIR)
	}
}

func (t *templateCmd) printFile(p string) error {
	if !score.FileExists(p) {
		return fmt.Errorf("no such template: %q", p)
	}

	fl, err := os.Open(p)

	if err != nil {
		return fmt.Errorf("can't open template %q: %s", p, err.Error())
	}

	defer fl.Close()
	io.Copy(os.Stdout, fl)
	return nil
}

func (t *templateCmd) printDir(dir string) error {
	fls, err := ioutil.ReadDir(muskel.USER_DIR)
	if err != nil {
		return fmt.Errorf("can't read user directory %q: %s", dir, err.Error())
	}

	if len(fls) == 0 {
		fmt.Fprintf(os.Stdout, "There are no template files stored inside %s\n\n", dir)
		return nil
	}

	fmt.Fprintf(os.Stdout, "The following template files are stored inside %s:\n\n", dir)

	for _, fl := range fls {
		fmt.Fprintf(os.Stdout, "%-24v (modified: %s)\n", fl.Name(), fl.ModTime().Format(time.RFC822))
	}
	return nil
}
