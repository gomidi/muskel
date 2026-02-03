package main

type Process struct {
	Program string
	Args    string
	PID     int
	killer  func()
}

func newProcess(program string, args string) *Process {
	return &Process{
		Program: program,
		Args:    args,
	}
}
