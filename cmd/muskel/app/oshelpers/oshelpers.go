package oshelpers

import (
	"fmt"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"
)

type Process struct {
	Program string
	Args    string
	PID     int
	killer  func()
}

func NewProcess(program string, args string) *Process {
	return &Process{
		Program: program,
		Args:    args,
	}
}

func AllVersionate(file string, v *version.Version) string {
	return fmt.Sprintf("%s_%v_%v_%v", file, v.Major, v.Minor, v.Patch)
}

func Open(file path.Local, async bool) {
	if async {
		go func() {
			OpenInDefaultProgram(file)
		}()
		return
	}
	OpenInDefaultProgram(file)
}
