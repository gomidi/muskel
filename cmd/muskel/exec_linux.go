// +build linux

package main

import (
	"os/exec"
	"syscall"

	"github.com/gen2brain/beeep"
	"gitlab.com/metakeule/config"
)

func (p *Process) Start() error {
	c := exec.Command("/bin/sh", "-c", "exec "+p.Program+" "+p.Args)
	err := c.Start()
	if err != nil {
		return err
	}

	p.PID = c.Process.Pid
	return nil
}

func (p *Process) Run() error {
	c := exec.Command("/bin/sh", "-c", "exec "+p.Program+" "+p.Args)
	_, err := c.CombinedOutput()
	if err != nil {
		return err
	}

	p.PID = c.Process.Pid
	return nil
}

func (p *Process) Kill() {
	syscall.Kill(p.PID, 9)
}

func defaultPlayCmd() [2]string {
	return [2]string{"audacious", "-1 -H -p -q $_file"}
}

/*
func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", "exec "+c)
}
*/

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
