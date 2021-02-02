package main

import (
	"fmt"
	"os"

	"gitlab.com/metakeule/config"
)

type portsCmd struct {
	*config.Config
}

var PORTS = portsCmd{}

func init() {
	PORTS.init()
}

func (s *portsCmd) init() {
	s.Config = CONFIG.MustCommand("ports", "show midi out ports").SkipAllBut()
}

func (p *portsCmd) run() error {
	drv, err := newDriver()
	if err != nil {
		return err
	}

	outs, err := drv.Outs()

	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "MIDI outputs")

	for _, out := range outs {
		fmt.Fprintf(os.Stdout, "[%v] %s\n", out.Number(), out.String())
	}

	return nil
}
