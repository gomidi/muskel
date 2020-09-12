// +build !windows,!linux

package main

import (
	"os/exec"
)

func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", c)
}

func alert(msg string, err error) {

}

func notify(msg, details string) {

}
