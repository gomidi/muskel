//go:build !windows && !linux
// +build !windows,!linux

package oshelpers

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"

	// "gitlab.com/gomidi/midi"
	// driver "gitlab.com/gomidi/midicatdrv"
	_ "gitlab.com/gomidi/midi/v2/drivers/midicatdrv"
)

/*
func newDriver() (midi.Driver, error) {
	return driver.New()
}
*/

func (p *Process) Start() error {
	cmd := exec.Command("/bin/sh", "-c", "exec "+p.Program+" "+p.Args)
	err := cmd.Start()
	if err != nil {
		return err
	}

	p.PID = cmd.Process.Pid
	return nil
}

func (p *Process) Run() error {
	cmd := exec.Command("/bin/sh", "-c", "exec "+p.Program+" "+p.Args)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	p.PID = cmd.Process.Pid
	return nil
}

func (p *Process) Kill() {
	syscall.Kill(p.PID, 9)
	time.Sleep(time.Second)
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

func Alert(msg string, err error) {

}

func Notify(msg, details string) {

}

func RunVersionated(file string, args []string) *exec.Cmd {
	return exec.Command(file, args...)
}

func Versionate(file string, v *version.Version) string {
	return AllVersionate(file, v)
}

func OpenInDefaultProgram(file path.Local) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("exec xdg-open '%s' 2>/dev/null", file.ToSystem()))
	return cmd.Run()
}
