//go:build !windows && !linux
// +build !windows,!linux

package main

import (
	"os/exec"
	"syscall"
	"time"

	"gitlab.com/golang-utils/version"
	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/midicatdrv"
)

func newDriver() (midi.Driver, error) {
	return driver.New()
}

func (p *Process) Start() error {
	p.cmd = exec.Command("/bin/sh", "-c", "exec "+p.Program+" "+p.Args)
	err := p.cmd.Start()
	if err != nil {
		return err
	}

	p.PID = p.cmd.Process.Pid
	return nil
}

func (p *Process) Run() error {
	p.cmd = exec.Command("/bin/sh", "-c", "exec "+p.Program+" "+p.Args)
	_, err := p.cmd.CombinedOutput()
	if err != nil {
		return err
	}

	p.PID = p.cmd.Process.Pid
	return nil
}

func (p *Process) Kill() {
	syscall.Kill(p.PID, 9)
	time.Sleep(time.Second)
}

func fluidsynthCmd() [2]string {
	return [2]string{"fluidsynth", "-i -n $_file"}
}

func audaciousCmd() [2]string {
	return [2]string{"audacious", "-1 -H -p -q $_file"}
}

func timidityCmd() [2]string {
	return [2]string{"timidity", "-V linear --noise-shaping=1 $_file"}
}

func defaultPlayCmd() [2]string {
	return fluidsynthCmd()
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

func versionate(file string, v *version.Version) string {
	return allVersionate(file, v)
}
