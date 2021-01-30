// +build !windows,!linux

package main

import (
	"os/exec"
	"syscall"

	"gitlab.com/metakeule/config"
)

func killCmd(c *exec.Cmd, pid int) {
	syscall.Kill(pid, 9)
}

func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", "exec "+c)
}

func alert(msg string, err error) {

}

func notify(msg, details string) {

}

func runVersionated(file string, args []string) *exec.Cmd {
	return exec.Command(file, args...)
}

func versionate(file string, v *config.Version) string {
	return config.Versionate(file, v)
}
