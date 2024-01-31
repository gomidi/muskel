package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/golang-utils/version"
	"gitlab.com/gomidi/muskel"
)

func checkMuskelVersion(version string, versionFile string, a *args) {
	srcdir := filepath.Dir(a.File.Get())
	v, err := muskel.ReadWDVersionFile(srcdir)
	if err == nil {
		if v.String() != version {
			name := versionate("muskel", v)
			fmt.Fprintf(os.Stderr, "This is version "+version+" of muskel and your "+versionFile+
				" points to version "+v.String()+
				".\nTo preserve compatibility with your musical notation, the binary \""+name+"\" will be called.\nIf you don't want this behavior or have no such older versioned muskel file, "+
				"\nremove the file \""+filepath.Join(srcdir, versionFile)+"\"\nor pass the \"--current\" option to let your file be parsed anyway.\n\n")

			cmd := runVersionated(name, os.Args[1:])
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
	muskel.WriteWDVersionFile(srcdir)
}

func allVersionate(file string, v *version.Version) string {
	// muskel_1_0_29.exe
	return fmt.Sprintf("%s_%v_%v_%v", file, v.Major, v.Minor, v.Patch)
}
