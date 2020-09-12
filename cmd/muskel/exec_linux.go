// +build linux

package main

import (
	"github.com/gen2brain/beeep"
	"os/exec"
)

func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", c)
}

func alert(msg string, err error) {
	//beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	beeep.Alert(msg, err.Error(), "assets/warning.png")
}

func notify(msg, details string) {
	beeep.Notify(msg, details, "assets/information.png")
}
