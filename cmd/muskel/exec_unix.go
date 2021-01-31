// +build !windows,!linux

package main

import (
	"os/exec"
	"syscall"
	"time"

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
	time.Sleep(time.Second)
}

func defaultPlayCmd() [2]string {
	return [2]string{"fluidsynth", "-i -n $_file"}
}

/*
func killCmd(c *exec.Cmd, pid int) {
	syscall.Kill(pid, 9)
}

func defaultPlayCmd() string {
	return "fluidsynth -i -q -n $_file"
}

func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", "exec "+c)
}
*/

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
