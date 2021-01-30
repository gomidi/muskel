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
	/*
		powershell.exe /C '$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = \"fluidsynth.exe\" ; $ProcessInfo.RedirectStandardError = $true ; $ProcessInfo.RedirectStandardOutput = $true; $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = \"-i -q -n main.mid\"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() ; echo $Process.Id'
	*/
	c := exec.Command("powershell.exe",
		"/Command",
		//`$Process = [Diagnostics.Process]::Start("`+p.Program+`.exe", "`+p.Args+`") ; echo $Process.Id `,
		`$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = "`+p.Program+`.exe" ; $ProcessInfo.RedirectStandardError = $true ; $ProcessInfo.RedirectStandardOutput = $true; $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = "`+p.Args+`"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() ; echo $Process.Id `,
	)
	out, err := c.CombinedOutput()
	if err != nil {
		return err
	}

	str := strings.Replace(string(out), "True", "", -1)
	p.PID, err = strconv.Atoi(strings.TrimSpace(str))
	return err
}

func (p *Process) Run() error {
	c := exec.Command("powershell.exe",
		"/Command",
		//`"`+p.Program+`.exe `+p.Args+`"`,
		p.Program+`.exe `+p.Args,
	)
	out, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	p.PID, err = strconv.Atoi(strings.TrimSpace(string(out)))
	return err
}

func (p *Process) Kill() {
	cmd := exec.Command("taskkill.exe", "/F", "/PID", fmt.Sprintf("%v", p.PID))
	_, _ = cmd.CombinedOutput()
}

func defaultPlayCmd() [2]string {
	//	return `C:\Program Files\fluidsynth\fluidsynth.exe -i -q -n $_file`
	//return "fluidsynth.exe -i -q -n $_file"
	return [2]string{"fluidsynth", "-i -n --quiet $_file"}
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
