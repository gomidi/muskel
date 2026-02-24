//go:build android
// +build android

package oshelpers

import (
	"fmt"

	//	"github.com/gen2brain/beeep"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"
	// driver "gitlab.com/gomidi/midicatdrv"
)

/*
func newDriver() (midi.Driver, error) {
	return driver.New()
}
*/

func (p *Process) Start() error {
	return fmt.Errorf("can't run processes on android")
}

func (p *Process) Run() error {
	return fmt.Errorf("can't run processes on android")
}

func (p *Process) Kill() {
	panic("can't run processes on android")
}

/*
func execCommand(c string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", "exec "+c)
}
*/

func Alert(msg string, err error) {
	//beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	//beeep.Alert(msg, err.Error(), "assets/warning.png")
}

func Notify(msg, details string) {
	//beeep.Notify(msg, details, "assets/information.png")
}

func Versionate(file string, v *version.Version) string {
	return AllVersionate(file, v)
}

func OpenInDefaultProgram(file path.Local) error {
	return fmt.Errorf("can't run processes on android")
}
