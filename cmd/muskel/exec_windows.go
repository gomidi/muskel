// +build windows

package main

import (
	"os/exec"

	"gitlab.com/metakeule/config"
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

func versionate(file string, v *config.Version) string {
	return config.Versionate(file+".exe", v)
}
