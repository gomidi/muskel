//go:build linux
// +build linux

package main

import (
	"os/exec"
	"syscall"
	"time"

	//	"github.com/gen2brain/beeep"
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

func vlcCmd() [2]string {
	return [2]string{"vlc", "--no-video --play-and-exit --no-playlist-tree --quiet -Idummy $_file"}
}

func fluidsynthCmd() [2]string {
	return [2]string{"fluidsynth", "-j -i -n $_file"}
}

func audaciousCmd() [2]string {
	return [2]string{"audacious", "-1 -H -p -q $_file"}
}

func timidityCmd() [2]string {
	return [2]string{"timidity", "-V linear --noise-shaping=1 $_file"}
}

func defaultPlayCmd() [2]string {
	return audaciousCmd()
}

/*
func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", "exec "+c)
}
*/

func alert(msg string, err error) {
	//beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	//beeep.Alert(msg, err.Error(), "assets/warning.png")
}

func notify(msg, details string) {
	//beeep.Notify(msg, details, "assets/information.png")
}

func runVersionated(file string, args []string) *exec.Cmd {
	return exec.Command(file, args...)
}

func versionate(file string, v *version.Version) string {
	return allVersionate(file, v)
}
