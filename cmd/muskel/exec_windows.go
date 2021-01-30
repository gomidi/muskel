// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"unsafe"

	"gitlab.com/metakeule/config"
	"golang.org/x/sys/windows"
)

// see https://gist.github.com/hallazzang/76f3970bfc949831808bbebc8ca15209

// We use this struct to retreive process handle(which is unexported)
// from os.Process using unsafe operation.
type process struct {
	Pid    int
	Handle uintptr
}

type ProcessExitGroup windows.Handle

func NewProcessExitGroup() (ProcessExitGroup, error) {
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, err
	}

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(
		handle,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info))); err != nil {
		return 0, err
	}

	return ProcessExitGroup(handle), nil
}

func (g ProcessExitGroup) Dispose() error {
	return windows.CloseHandle(windows.Handle(g))
}

func (g ProcessExitGroup) AddProcess(p *os.Process) error {
	return windows.AssignProcessToJobObject(
		windows.Handle(g),
		windows.Handle((*process)(unsafe.Pointer(p)).Handle))
}

func (p *Process) Start() error {
	/*
		powershell.exe /C '$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = \"fluidsynth.exe\" ; $ProcessInfo.RedirectStandardError = $true ; $ProcessInfo.RedirectStandardOutput = $true; $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = \"-i -q -n main.mid\"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() ; echo $Process.Id'
	*/
	g, err := NewProcessExitGroup()
	if err != nil {
		return err
	}

	c := exec.Command("powershell.exe",
		"/Command",
		//`$Process = [Diagnostics.Process]::Start("`+p.Program+`.exe", "`+p.Args+`") ; echo $Process.Id `,
		//`$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = "`+p.Program+`.exe" ; $ProcessInfo.RedirectStandardError = $true ; $ProcessInfo.RedirectStandardOutput = $true; $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = "`+p.Args+`"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() ; echo $Process.Id `,
		`$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = "`+p.Program+`.exe" ;  $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = "`+p.Args+`"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() | Out-Null ; echo $Process.Id `,
	)

	err = c.Start()

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	err = g.AddProcess(c.Process)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	p.killer = func() {
		g.Dispose()
	}

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
	//cmd := exec.Command("taskkill.exe", "/F", "/PID", fmt.Sprintf("%v", p.PID))
	//_, _ = cmd.CombinedOutput()
	p.killer()
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
