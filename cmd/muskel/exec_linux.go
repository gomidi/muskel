// +build linux

package main

import (
	"os/exec"

	"github.com/gen2brain/beeep"
	"gitlab.com/metakeule/config"
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

func runVersionated(file string, args []string) *exec.Cmd {
	return exec.Command(file, args...)
}

func versionate(file string, v *config.Version) string {
	return config.Versionate(file, v)
}
