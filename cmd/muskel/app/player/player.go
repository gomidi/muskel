package player

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/smfplayer"
)

type Player struct {
	PlayToPort   bool
	Enabled      bool
	RenderToWave bool
	PortOut      drivers.Out
	PortRegEx    *regexp.Regexp
	SMFFile      path.Local
	WavFile      path.Local

	playCh    chan bool
	stoppedCh chan bool
	toggleCh  chan bool

	stopPlayer chan bool
	wasStopped chan bool
}

func New() *Player {
	return &Player{
		playCh:    make(chan bool, 10),
		toggleCh:  make(chan bool, 10),
		stoppedCh: make(chan bool, 10),

		stopPlayer: make(chan bool, 10),
		wasStopped: make(chan bool, 10),
		PlayToPort: false,

		PortRegEx: regexp.MustCompile("port" + regexp.QuoteMeta(":") + "([0-9]+)"),
	}
}

func (me *Player) Play() {
	me.playCh <- true
}

func (me *Player) StopFunc() func() {
	return func() {
		me.stopPlayer <- true
		<-me.wasStopped
	}
}

func (me *Player) Stop() {
	me.playCh <- false
	<-me.stoppedCh
}
func (me *Player) Toggle() {
	me.toggleCh <- true
}

func (me *Player) converttoWav() error {
	fmt.Println("converttoWav")
	fsys, err := rootfs.New()
	if err != nil {
		return err
	}

	sp := smfplayer.New(int32(44100))

	_fsys, file, err := smfplayer.DefaultSF2()
	if err != nil {
		return err
	}
	err = sp.LoadSF2(_fsys, file)
	if err != nil {
		return err
	}

	err = sp.SynthesizeSMF(fsys, me.SMFFile.RootRelative())
	if err != nil {
		return err
	}

	return sp.WriteFile(fsys, me.WavFile.RootRelative())
}

func (me *Player) ConvertAndPlay(out *smfplayer.AudioOut) error {
	fmt.Println("ConvertAndPlay")
	if me.RenderToWave {
		return me.converttoWav()
	}
	fsys, err := rootfs.New()
	if err != nil {
		return err
	}

	fmt.Println("Play as audio")
	sp := smfplayer.New(int32(44100))

	_fsys, file, err := smfplayer.DefaultSF2()
	if err != nil {
		return err
	}
	err = sp.LoadSF2(_fsys, file)
	if err != nil {
		return err
	}

	err = sp.SynthesizeSMF(fsys, me.SMFFile.RootRelative())
	if err != nil {
		return err
	}

	return sp.Play(out)
}

/*
out, err := smfplayer.OpenAudioOut(me.SampleRate.Val)

	if err != nil {
		return err
	}

	defer out.Close()

	var player *smfplayer.Player

	go func() {
		<-SIGNAL_CHANNEL
		fmt.Println("\n--interrupted!")
		player.StopPlaying()
		out.Close()
		os.Exit(1)
	}()

	for range me.PlayingTimes.Val {
		player, _, err = me.setupPlayer()
		if err != nil {
			return err
		}

		err = player.Play(out)

		if err != nil {
			return err
		}

		time.Sleep(time.Second)
		fmt.Print(".")
	}
*/

func (me *Player) ConvertAndPlay2(stopPortPlayer chan bool, stoppedPortPlayer chan bool, out *smfplayer.AudioOut) error {
	fmt.Println("ConvertAndPlay2")
	if me.RenderToWave {
		return me.converttoWav()
	}
	fsys, err := rootfs.New()
	if err != nil {
		return err
	}

	fmt.Println("Play as audio")
	sp := smfplayer.New(int32(44100))

	_fsys, file, err := smfplayer.DefaultSF2()
	if err != nil {
		return err
	}
	err = sp.LoadSF2(_fsys, file)
	if err != nil {
		return err
	}

	err = sp.SynthesizeSMF(fsys, me.SMFFile.RootRelative())
	if err != nil {
		return err
	}

	err = sp.Play(out)

	if err != nil {
		return err
	}

	fmt.Println("start playing")

	for {
		select {
		case <-stopPortPlayer:
			sp.StopPlaying()
			stoppedPortPlayer <- true
			me.wasStopped <- true
			return nil
		default:
		}
	}
}

func (me *Player) PlayAsService(out *smfplayer.AudioOut) {
	fmt.Println("PlayAsService")

	for {
		select {
		case <-me.toggleCh:
			me.ConvertAndPlay(out)
		case <-me.stopPlayer:
			me.wasStopped <- true
			fmt.Println("player stopped")
			return
		case shouldPlay := <-me.playCh:
			_ = shouldPlay
			me.ConvertAndPlay(out)
			me.stoppedCh <- true
		}
	}
}

func (me *Player) PlayThroughPort() {
	fmt.Println("playThroughPort")
	var cmd *smf.TracksReader

	var stopPortPlayer = make(chan bool, 1)
	var stoppedPortPlayer = make(chan bool, 1)
	var mx sync.Mutex

	var startPlayer = func() {
		mx.Lock()
		stopPortPlayer = make(chan bool, 1)
		stoppedPortPlayer = make(chan bool, 1)

		cmd = smf.ReadTracks(path.ToSystem(me.SMFFile))
		//cmd, err = player.SMF(p.outFile)

		go func() {
			err := cmd.Play(me.PortOut)
			if err != nil {
				fmt.Println("ERROR: could not play")
			}
			//cmd.PlayAll(p.portOut, stopPortPlayer, stoppedPortPlayer)
			mx.Unlock()
		}()

	}

	for {
		select {
		case <-me.toggleCh:
			if cmd != nil {
				stopPortPlayer <- true
				<-stoppedPortPlayer
				mx.Lock()
				cmd = nil
				mx.Unlock()
			} else {
				startPlayer()
			}
		case <-me.stopPlayer:
			me.PortOut.Close()
			if cmd != nil {
				stopPortPlayer <- true
				<-stoppedPortPlayer
				mx.Lock()
				cmd = nil
				mx.Unlock()
			}
			me.wasStopped <- true
			return
		case shouldPlay := <-me.playCh:
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
				me.stoppedCh <- true
			}
		}
	}
}

func (me *Player) PlayOnce(stopPortPlayer chan bool, stoppedPortPlayer chan bool, out *smfplayer.AudioOut) {
	if me.PlayToPort {
		trks := smf.ReadTracks(path.ToSystem(me.SMFFile))
		go func() {

			// stops playing, if  either trks.StopPlaying() was called, or the playing is finished
			err := trks.Play(me.PortOut)
			if err != nil {
				fmt.Println("ERROR: could not play")
			}

			time.Sleep(30 * time.Millisecond)

			// to be sure, stop all "hanging" notes
			for ch := range 16 {
				//fmt.Printf("sending all notes off to channel: %v\n", ch)
				me.PortOut.Send(midi.ControlChange(uint8(ch), midi.AllNotesOff, midi.Off).Bytes())
			}

			//fmt.Println("finished playing to port")
			//fmt.Println("closing port (player)")

			//fmt.Println("wasStopped sent")
			// finish the for select loop below
			me.wasStopped <- true
			//fmt.Println("stoppedPortPlayer")
			// tell the outer app that we finished
			stoppedPortPlayer <- true
		}()

		for {
			select {
			case <-me.wasStopped:
				return
			case <-stopPortPlayer:
				trks.StopPlaying()
				return
			}
		}
	} else {
		err := me.ConvertAndPlay2(stopPortPlayer, stoppedPortPlayer, out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while playing %s", err)
		}
	}
}
