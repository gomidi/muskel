package main

import (
	"fmt"
	"os"

	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/gomidi/midi/v2"
)

type portsCmd struct {
	*config.Config
}

var PORTS = portsCmd{}

func init() {
	PORTS.init()
}

func (s *portsCmd) init() {
	s.Config = CONFIG.Command("ports", "show midi out ports").SkipAllBut()
}

func (p *portsCmd) run() error {
	/*
		drv, err := newDriver()
		if err != nil {
			return err
		}
	*/

	outs := midi.GetOutPorts()
	/*
		outs, err := drv.Outs()

		if err != nil {
			return err
		}
	*/

	fmt.Fprintln(os.Stdout, "MIDI outputs")

	for _, out := range outs {
		fmt.Fprintf(os.Stdout, "[%v] %s\n", out.Number(), out.String())
	}

	return nil
}
