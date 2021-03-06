package main

import (
	"gitlab.com/metakeule/config"
)

type playCmd struct {
	*config.Config
	Program config.StringGetter
}

var PLAY = playCmd{}

func init() {
	PLAY.init()
}

func (p *playCmd) init() {
	p.Config = CONFIG.MustCommand("play", "play a muskel file via smfplayer (timidity,fluidsynth or audacity) or directly to midi out ports").Skip("watch").Skip("dir").Skip("sleep")
	p.Program = p.NewString("cmd", "command to execute when playing (fluidsynth,timidity,audacious,auto or midi-ports (port:4) or custom (pass $_file variable))", config.Default("auto"))
}
