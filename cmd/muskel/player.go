package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/smf"
	//"gitlab.com/gomidi/midi/player"
)

type Player struct {
	playCh    chan bool
	stoppedCh chan bool
	toggleCh  chan bool

	stopPlayer    chan bool
	playerStopped chan bool

	program [2]string

	playToPort bool
	portOut    drivers.Out

	cmdMaps map[string][2]string

	portRegEx *regexp.Regexp

	outFile path.Local

	playerEnabled bool
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
			"timidity":   timidityCmd(),
			"audacious":  audaciousCmd(),
			"fluidsynth": fluidsynthCmd(),
			"vlc":        vlcCmd(),
			"auto":       defaultPlayCmd(),
		},

		portRegEx: regexp.MustCompile("port" + regexp.QuoteMeta(":") + "([0-9]+)"),
	}

	err = p.setupProgram()
	return
}

func (ps *Player) mkPlayCmdString(cmd string, outfile path.Local) {
	cm := strings.TrimSpace(cmd)
	//fmt.Printf("cmd : %q\n", cm)
	var cmm [2]string
	if cc, has := ps.cmdMaps[cm]; has {
		cmm = cc
	} else {
		cmm_ := strings.SplitN(cm, " ", 2)
		cmm[0] = cmm_[0]
		cmm[1] = cmm_[1]
	}

	cmm[1] = strings.ReplaceAll(cmm[1], "$_file", ps.outFile.String())
	ps.program = cmm
}

func (p *Player) normalizeOutFile(a *args) {
	inFile := a.InFile.Get()
	p.outFile = inFile
	if a.OutFile.IsSet() {
		p.outFile = a.OutFile.Get()
	}

	if a.TrackFiles.Get() {
		p.outFile = a.OutFile.Get().Dir().Join("%s")
	}

	if extIdx := strings.LastIndex(p.outFile.String(), "."); extIdx > 0 && extIdx+1 < len(p.outFile.String()) {
		switch p.outFile.String()[extIdx:] {
		case ".mid", ".midi", ".MID", ".MIDI":
		default:
			p.outFile = path.MustLocal(p.outFile.String()[:extIdx] + ".mid")
		}
	} else {
		p.outFile = path.MustLocal(p.outFile.String() + ".mid")
	}
}

func (ps *Player) setupProgram() error {
	var cmd string

	ps.normalizeOutFile(ARGS)

	if CONFIG.ActiveCommand() != PLAY.Config && CONFIG.ActiveCommand() != SERVER.Config {
		return nil
	}

	ps.playerEnabled = true

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
		/*
			drv, err := newDriver()
			if err != nil {
				return fmt.Errorf("can't open midi driver")
			}
		*/
		outs := midi.GetOutPorts()
		if len(outs)-1 < p {
			return fmt.Errorf("can't open midi out port %v", p)
		}
		//ps.portOut, err = midi.OpenOut(drv, p, "")
		ps.portOut = outs[p]
		if err != nil {
			return fmt.Errorf("can't open midi out port %v", p)
		}
	} else {
		ps.mkPlayCmdString(cmd, ps.outFile)
		fmt.Printf("%s %s\n", ps.program[0], ps.program[1])
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
			fmt.Println("player stopped")
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
	var cmd *smf.TracksReader

	var stopPortPlayer = make(chan bool, 1)
	var stoppedPortPlayer = make(chan bool, 1)
	var mx sync.Mutex

	var startPlayer = func() {
		mx.Lock()
		stopPortPlayer = make(chan bool, 1)
		stoppedPortPlayer = make(chan bool, 1)

		cmd = smf.ReadTracks(path.ToSystem(p.outFile))
		//cmd, err = player.SMF(p.outFile)

		go func() {
			err := cmd.Play(p.portOut)
			if err != nil {
				fmt.Println("ERROR: could not play")
			}
			//cmd.PlayAll(p.portOut, stopPortPlayer, stoppedPortPlayer)
			mx.Unlock()
		}()

	}

	for {
		select {
		case <-p.toggleCh:
			if cmd != nil {
				stopPortPlayer <- true
				<-stoppedPortPlayer
				mx.Lock()
				cmd = nil
				mx.Unlock()
			} else {
				startPlayer()
			}
		case <-p.stopPlayer:
			p.portOut.Close()
			if cmd != nil {
				stopPortPlayer <- true
				<-stoppedPortPlayer
				mx.Lock()
				cmd = nil
				mx.Unlock()
			}
			p.playerStopped <- true
			return
		case shouldPlay := <-p.playCh:
			if shouldPlay {
				if cmd != nil {
					stopPortPlayer <- true
					<-stoppedPortPlayer
					mx.Lock()
					cmd = nil
					mx.Unlock()
				}
				startPlayer()
			} else {
				if cmd != nil {
					stopPortPlayer <- true
					<-stoppedPortPlayer
					mx.Lock()
					cmd = nil
					mx.Unlock()
				}
				p.stoppedCh <- true
			}
		}
	}
}

func (p *Player) playOnce(stopPortPlayer chan bool, stoppedPortPlayer chan bool) {
	if p.playToPort {
		pl := smf.ReadTracks(path.ToSystem(p.outFile))
		/*
			pl, err := player.SMF(p.outFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: could not create smfplayer")
			}
		*/
		pl.Play(p.portOut)
		//pl.PlayAll(p.portOut, stopPortPlayer, stoppedPortPlayer)
	} else {
		cmd := newProcess(p.program[0], p.program[1])
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while running %s %s: %v", cmd.Program, cmd.Args, err)
		}
	}
}
