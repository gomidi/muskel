package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel"
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
	s.Config = CONFIG.Command("template", "show global template files").SkipAllBut()
	s.File = s.LastString("templatefile", "file name of the template file that should be shown. If no file is given, the list of available files is shown.")
}

func (t *templateCmd) run() error {
	if t.File.IsSet() && t.File.Get() != "" {
		return t.printFile(muskel.USER_DIR.Join(t.File.Get() + muskel.FILE_EXTENSION))
	} else {
		return t.printDir(muskel.USER_DIR)
	}
}

func (t *templateCmd) printFile(p path.Local) error {
	if !rootfs.Exists(p) {
		return fmt.Errorf("no such template: %q", p)
	}

	fl, err := rootfs.OpenReader(p)

	if err != nil {
		return fmt.Errorf("can't open template %q: %s", p, err.Error())
	}

	defer fl.Close()
	io.Copy(os.Stdout, fl)
	return nil
}

func (t *templateCmd) printDir(dir path.Local) error {
	fls, err := rootfs.ReadDirPaths(dir)
	if err != nil {
		return fmt.Errorf("can't read user directory %q: %s", dir, err.Error())
	}

	if len(fls) == 0 {
		fmt.Fprintf(os.Stdout, "There are no template files stored inside %s\n\n", dir)
		return nil
	}

	fmt.Fprintf(os.Stdout, "The following template files are stored inside %s:\n\n", dir)

	for _, fl := range fls {
		name := fl.String()
		var tistr string
		ti, err := rootfs.ModTime(fl)
		if err == nil {
			tistr = fmt.Sprintf("(modified: %s)", ti.Format(time.RFC822))
		}
		fmt.Fprintf(os.Stdout, "%-24v %s\n", name, tistr)
	}
	return nil
}
