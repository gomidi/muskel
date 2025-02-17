package main

import (
	"fmt"
	"os"
	"time"

	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/golang-utils/fs/path"
)

func watch(conv *converter, a *args) error {
	if a.WatchDir.Get() {
		fmt.Printf("watching %q\n", conv.dir)
	} else {
		fmt.Printf("watching %q\n", a.InFile.Get())
	}

	r := runfunc.New(
		path.ToSystem(conv.dir),
		conv.mkcallback(),
		runfunc.Sleep(time.Millisecond*time.Duration(int(a.SleepingTime.Get()))),
	)

	errors := make(chan error, 1)
	stopped, err := r.Run(errors)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			select {
			case e := <-errors:
				fmt.Fprintf(os.Stderr, "error: %s\n", e)
			}
		}
	}()

	// ctrl+c
	<-SIGNAL_CHANNEL
	fmt.Println("\n--interrupted!")
	stopped.Kill()

	os.Exit(0)
	return nil
}
