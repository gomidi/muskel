//go:build android
// +build android

package main

import (
	"fmt"
	"os"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel"
)

func checkMuskelVersion(version string, versionFile string, a *args) {
	srcdir := a.InFile.Val.Dir()
	v, err := muskel.ReadWDVersionFile(srcdir)
	if err == nil {
		if v.String() != version {
			name := versionate("muskel", v)
			fmt.Fprint(os.Stderr, "This is version "+version+" of muskel and your "+versionFile+
				" points to version "+v.String()+
				".\nTo preserve compatibility with your musical notation, run \""+name+"\" instead.\nIf you don't have no such older versioned muskel file, "+
				"\nremove the file \""+srcdir.Join(versionFile).String()+"\"\nor pass the \"--current\" option to let your file be parsed anyway.\n\n")

			os.Exit(1)
		}
	}
	muskel.WriteWDVersionFile(path.ToSystem(srcdir))
}
