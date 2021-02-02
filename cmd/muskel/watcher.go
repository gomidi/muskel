package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/metakeule/observe/lib/runfunc"
)

func watch(conv *converter, a *args) error {
	if a.WatchDir.Get() {
		fmt.Printf("watching %q\n", conv.dir)
	} else {
		fmt.Printf("watching %q\n", a.File.Get())
	}

	r := runfunc.New(
		conv.dir,
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
			default:
				runtime.Gosched()
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
