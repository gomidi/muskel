//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"

	//"strconv"
	//"strings"

	"syscall"

	"unsafe"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"

	//	"gitlab.com/gomidi/midi"

	//	driver "gitlab.com/gomidi/rtmididrv"
	//	driver "gitlab.com/gomidi/midicatdrv"
	_ "gitlab.com/gomidi/midi/v2/drivers/midicatdrv"
	"golang.org/x/sys/windows"
)

/*
func newDriver() (midi.Driver, error) {
	return driver.New()
}
*/

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

	p.cmd = exec.Command("powershell.exe",
		"/Command",
		//`$Process = [Diagnostics.Process]::Start("`+p.Program+`.exe", "`+p.Args+`") ; echo $Process.Id `,
		//`$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = "`+p.Program+`.exe" ; $ProcessInfo.RedirectStandardError = $true ; $ProcessInfo.RedirectStandardOutput = $true; $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = "`+p.Args+`"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() ; echo $Process.Id `,
		`$ProcessInfo = New-Object System.Diagnostics.ProcessStartInfo; $ProcessInfo.FileName = "`+p.Program+`" ;  $ProcessInfo.UseShellExecute = $false; $ProcessInfo.Arguments = "`+p.Args+`"; $Process = New-Object System.Diagnostics.Process; $Process.StartInfo = $ProcessInfo ; $Process.Start() | Out-Null ; echo $Process.Id `,
	)

	err = p.cmd.Start()

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	err = g.AddProcess(p.cmd.Process)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	p.killer = func() {
		g.Dispose()
	}

	return nil

}

func (p *Process) Run() error {
	p.cmd = exec.Command("powershell.exe",
		"/Command",
		//`"`+p.Program+`.exe `+p.Args+`"`,
		p.Program+` `+p.Args,
	)

	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP, //  .CREATE_NEW_CONSOLE,
	}

	out, err := p.cmd.CombinedOutput()
	if err != nil {
		return err
	}

	_ = out

	/*
		outStr := strings.TrimSpace(string(out))

		if outStr == "" {
			p.PID = -1
			return nil
		}
	*/

	//p.PID, err = strconv.Atoi(outStr)
	return err
}

func (p *Process) Kill() {
	//cmd := exec.Command("taskkill.exe", "/F", "/PID", fmt.Sprintf("%v", p.PID))
	//_, _ = cmd.CombinedOutput()
	p.killer()
}

func vlcCmd() [2]string {
	return [2]string{"vlc.exe", "--no-video --play-and-exit --no-playlist-tree --quiet -Idummy $_file"}
}

//vlc.exe --no-video --play-and-exit --no-playlist-tree --quiet -Idummy main.mid

func fluidsynthCmd() [2]string {
	return [2]string{"fluidsynth.exe", "-i -n --quiet $_file"}
}

func audaciousCmd() [2]string {
	return [2]string{"audacious.exe", "-1 -H -p -q $_file"}
}

func timidityCmd() [2]string {
	return [2]string{"timidity.exe", "--quiet -V linear --noise-shaping=1 $_file"}
}

func defaultPlayCmd() [2]string {
	//	return `C:\Program Files\fluidsynth\fluidsynth.exe -i -q -n $_file`
	return vlcCmd()
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

func versionate(file string, v *version.Version) string {
	return allVersionate(file, v) + ".exe"
}

func openInDefaultProgram(file path.Local) error {
	cmd := exec.Command("cmd.exe", "start", fmt.Sprintf(`"%s"`, file.ToSystem()))
	return cmd.Run()
}
