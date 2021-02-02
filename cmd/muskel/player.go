package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/player"
)

type Player struct {
	playCh    chan bool
	stoppedCh chan bool
	toggleCh  chan bool

	stopPlayer    chan bool
	playerStopped chan bool

	program [2]string

	playToPort bool
	portOut    midi.Out

	cmdMaps map[string][2]string

	portRegEx *regexp.Regexp

	outFile string
}

func newPlayer() (p *Player, err error) {
	p = &Player{
		playCh:    make(chan bool, 10),
		toggleCh:  make(chan bool, 10),
		stoppedCh: make(chan bool, 10),

		stopPlayer:    make(chan bool, 10),
		playerStopped: make(chan bool, 10),
		playToPort:    false,

		cmdMaps: map[string][2]string{
			"timidity":   [2]string{"timidity", "--quiet -V linear --noise-shaping=1 $_file"},
			"audacious":  [2]string{"audacious", "-1 -H -p -q $_file"},
			"fluidsynth": [2]string{"fluidsynth", "-i -n $_file"},
			"auto":       defaultPlayCmd(),
		},

		portRegEx: regexp.MustCompile("port" + regexp.QuoteMeta(":") + "([0-9]+)"),
	}

	err = p.setupProgram()
	return
}

func (ps *Player) setupProgram() error {
	var cmd string

	if CONFIG.ActiveCommand() == PLAY.Config {
		cmd = PLAY.Program.Get()
	}

	if CONFIG.ActiveCommand() == SERVER.Config {
		cmd = SERVER.PlayProgram.Get()
	}

	if ps.portRegEx.MatchString(cmd) {
		ps.playToPort = true
		m := ps.portRegEx.FindSubmatch([]byte(cmd))
		p, err := strconv.Atoi(string(m[1]))
		if err != nil {
			return fmt.Errorf("invalid port (not a number)")
		}
		fmt.Printf("out port: %v\n", p)
		drv, err := newDriver()
		if err != nil {
			return fmt.Errorf("can't open midi driver")
		}
		ps.portOut, err = midi.OpenOut(drv, p, "")
		if err != nil {
			return fmt.Errorf("can't open midi out port %v", p)
		}

	}
	return nil
}

func (p *Player) playWithProgram() {
	var cmd *Process
	var mx sync.Mutex

	var startProc = func() {
		mx.Lock()
		go func() {
			err := cmd.Start()
			mx.Unlock()
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR while running %s %s: %v", cmd.Program, cmd.Args, err)
			}
		}()
	}

	for {
		select {
		case <-p.toggleCh:
			if cmd != nil {
				cmd.Kill()
				cmd = nil
			} else {
				cmd = newProcess(p.program[0], p.program[1])
				startProc()
			}
		case <-p.stopPlayer:
			if cmd != nil {
				cmd.Kill()
				cmd = nil
			}
			p.playerStopped <- true
			return
		case shouldPlay := <-p.playCh:
			if shouldPlay {
				if cmd != nil {
					cmd.Kill()
					cmd = nil
				}
				cmd = newProcess(p.program[0], p.program[1])
				startProc()
			} else {
				if cmd != nil {
					cmd.Kill()
					cmd = nil
				}
				p.stoppedCh <- true
			}
		}
	}
}

func (p *Player) playThroughPort() {
	var cmd *player.Player

	var stopPortPlayer = make(chan bool, 1)
	var stoppedPortPlayer = make(chan bool, 1)

	for {
		select {
		case <-p.toggleCh:
			if cmd != nil {
				stopPortPlayer <- true
				<-stoppedPortPlayer
				cmd = nil
			} else {
				var err error
				cmd, err = player.SMF(p.outFile)
				if err != nil {
					fmt.Println("ERROR: could not create smfplayer")
				} else {
					go func() {
						cmd.PlayAll(p.portOut, stopPortPlayer, stoppedPortPlayer)
					}()
				}
			}
		case <-p.stopPlayer:
			if cmd != nil {
				stopPortPlayer <- true
				<-stoppedPortPlayer
				cmd = nil
			}
			p.playerStopped <- true
			return
		case shouldPlay := <-p.playCh:
			if shouldPlay {
				if cmd != nil {
					stopPortPlayer <- true
					<-stoppedPortPlayer
					cmd = nil
				}
				var err error
				cmd, err = player.SMF(p.outFile)
				if err != nil {
					fmt.Println("ERROR: could not create smfplayer")
				} else {
					go func() {
						cmd.PlayAll(p.portOut, stopPortPlayer, stoppedPortPlayer)
					}()
				}
			} else {
				if cmd != nil {
					stopPortPlayer <- true
					<-stoppedPortPlayer
					cmd = nil
				}
				p.stoppedCh <- true
			}
		}
	}
}

func (p *Player) playOnce(stopPortPlayer chan bool, stoppedPortPlayer chan bool) {
	if p.playToPort {
		pl, err := player.SMF(p.outFile)
		if err != nil {
			fmt.Println("ERROR: could not create smfplayer")
		}

		pl.PlayAll(p.portOut, stopPortPlayer, stoppedPortPlayer)
	} else {
		cmd := newProcess(p.program[0], p.program[1])
		fmt.Fprintf(os.Stdout, "%s %s\n", cmd.Program, cmd.Args)
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while running %s %s: %v", cmd.Program, cmd.Args, err)
		}
	}
}
