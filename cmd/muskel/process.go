package main

import (
	"os/exec"
)

type Process struct {
	Program string
	Args    string
	PID     int
	cmd     *exec.Cmd
	killer  func()
}

func newProcess(program string, args string) *Process {
	return &Process{
		Program: program,
		Args:    args,
	}
}
