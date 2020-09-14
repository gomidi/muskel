// +build windows

package main

import (
	"os/exec"

	"gitlab.com/gomidi/muskel"
)

func execCommand(c string) *exec.Cmd {
	return exec.Command("cmd.exe", "/C", c)
}

func alert(msg string, err error) {

}

func notify(msg, details string) {

}

func runVersionated(file string, args []string) *exec.Cmd {
	return exec.Command(file, args...)
}

func versionate(file string, v *muskel.Version) string {
	return muskel.Versionate(file+".exe", v)
}
