package main

import (
	"os/exec"
)

func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", c)
}
