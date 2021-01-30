// +build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"gitlab.com/metakeule/config"
)

func (p *Process) Start() error {
	c := exec.Command("powershell.exe",
		"/Command",
		`$Process = [Diagnostics.Process]::Start("`+p.Program+`.exe", "`+p.Args+`") ; echo $Process.Id `,
	)
	_, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	p.PID, err = strconv.Atoi(strings.TrimSpaces(string(out)))
	return err
}

func (p *Process) Run() error {
	c := exec.Command("powershell.exe",
		"/Command",
		`$Process = [Diagnostics.Process]::Start("`+p.Program+`.exe", "`+p.Args+`") ; echo $Process.Id `,
	)
	_, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	p.PID, err = strconv.Atoi(strings.TrimSpaces(string(out)))
	return err
}

func (p *Process) Kill() {
	cmd := exec.Command("taskkill.exe", "/F", "/PID", fmt.Sprintf("%v", pid))
	_, _ = cmd.CombinedOutput()
}

func defaultPlayCmd() [2]string {
	//	return `C:\Program Files\fluidsynth\fluidsynth.exe -i -q -n $_file`
	//return "fluidsynth.exe -i -q -n $_file"
	return [2]string{"fluidsynth", "-i -n $_file"}
}

/*
func execCommand(c string) *exec.Cmd {
	//return exec.Command("powershell.exe", "/Command",  `$Process = [Diagnostics.Process]::Start("` + c + `") ; echo $Process.Id `)
	return exec.Command("powershell.exe", "/Command", `$Process = [Diagnostics.Process]::Start("fluidsynth.exe", "-i -q -n $_file") ; echo $Process.Id `)
	//return exec.Command("cmd.exe", "/C", c)
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
	return config.Versionate(file+".exe", v)
}
