// +build windows

package main

import (
	"fmt"
	"os/exec"

	"gitlab.com/metakeule/config"
)

func killCmd(c *exec.Cmd, pid int) {
	//	c.Process.Kill()
	cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%v", pid))
	_, _ = cmd.CombinedOutput()
}

func defaultPlayCmd() string {
	//	return `C:\Program Files\fluidsynth\fluidsynth.exe -i -q -n $_file`
	return "fluidsynth.exe -i -q -n $_file"
}

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
