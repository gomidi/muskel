// +build !windows,!linux

package main

import (
	"os/exec"

	"gitlab.com/gomidi/muskel"
)

func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", c)
}

func alert(msg string, err error) {

}

func notify(msg, details string) {

}

func runVersionated(file string, args []string) *exec.Cmd {
	return exec.Command(file, args...)
}

func versionate(file string, v *muskel.Version) string {
	return muskel.Versionate(file, v)
}
