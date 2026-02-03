//go:build android
// +build android

package main

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

func versionate(file string, v *version.Version) string {
	return allVersionate(file, v)
}

func openInDefaultProgram(file path.Local) error {
	return fmt.Errorf("can't run processes on android")
}
