//go:build !android
// +build !android

package oshelpers

import (
	"fmt"
	"os"
	"os/exec"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel"
)

func CheckMuskelVersion(version string, versionFile string, srcdir path.Local) {
	v, err := muskel.ReadWDVersionFile(srcdir)
	if err == nil {
		if v.String() != version {
			name := Versionate("muskel", v)
			fmt.Fprint(os.Stderr, "This is version "+version+" of muskel and your "+versionFile+
				" points to version "+v.String()+
				".\nTo preserve compatibility with your musical notation, the binary \""+name+"\" will be called.\nIf you don't want this behavior or have no such older versioned muskel file, "+
				"\nremove the file \""+srcdir.Join(versionFile).String()+"\"\nor pass the \"--current\" option to let your file be parsed anyway.\n\n")

			cmd := RunVersionated(name, os.Args[1:])
			cmd.Dir, _ = os.Getwd()
			cmd.Env = os.Environ()
			var out []byte
			out, err = cmd.CombinedOutput()
			if err != nil {
				if _, notExeced := err.(*exec.Error); notExeced {
					fmt.Fprintf(os.Stderr, "Error: could not find %q binary\n", name)
					os.Exit(1)
				}
				fmt.Fprintln(os.Stderr, string(out))
				os.Exit(1)
			}
			fmt.Fprintln(os.Stdout, string(out))
			os.Exit(0)
		}
	}
	muskel.WriteWDVersionFile(path.ToSystem(srcdir))
}
