package main

/*
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/smfimport"
)

type importCmd struct {
	*config.Config
	Source          config.StringGetter
	OutFile         config.StringGetter
	MonoTracks      config.StringGetter
	KeysTracks      config.StringGetter
	SplitTypes      config.BoolGetter
	Quantize        config.BoolGetter
	SkipVelocity    config.BoolGetter
	SkipMicroTiming config.BoolGetter
}

var IMPORT = &importCmd{}

func init() {
	IMPORT.init()
}

func (s *importCmd) init() {
	s.Config = CONFIG.Command("import", "convert a Standard MIDI file to a muskel file").SkipAllBut("watch", "dir")
	s.Source = s.LastString("midifile", "the source SMF/MIDI file.", config.Required())
	s.OutFile = s.String("out", "the target muskel file.", config.Shortflag('o'), config.Required())
	s.MonoTracks = s.String("mono", "tracks that are considered to be monophon (e.g. 0,4,5)", config.Shortflag('m'))
	s.KeysTracks = s.String("keys", "tracks that should keep the key numbers instead of converting to note names (for drums keys, sample keys or keyswitches) (e.g. 0,4,5)", config.Shortflag('k'))
	s.SplitTypes = s.Bool("split", "split tracks by MIDI event types", config.Default(false), config.Shortflag('s'))
	s.Quantize = s.Bool("quantize", "quantize, based on a dual 16th/8th tripplets grid with auto track offset detection", config.Shortflag('q'))
	s.SkipVelocity = s.Bool("skipvelo", "do not import velocity, use standard velocity instead")
	s.SkipMicroTiming = s.Bool("skipmicro", "do not track micro timing")
}

func (i *importCmd) run(a *args) error {

	if a.Watch.IsSet() {
		return i.watch(a)
	}

	return i.runImport(i.Source.Get(), i.OutFile.Get())
}

func (i *importCmd) runImport(srcFile, outFile string) error {

	var opts []smfimport.Option

	if i.MonoTracks.IsSet() {
		monoTracks := convertArgStringToInts(i.MonoTracks.Get())
		if len(monoTracks) > 0 {
			opts = append(opts, smfimport.MonoTracks(monoTracks...))
		}
	}

	if i.KeysTracks.IsSet() {
		keysTracks := convertArgStringToInts(i.KeysTracks.Get())
		if len(keysTracks) > 0 {
			opts = append(opts, smfimport.KeysTracks(keysTracks...))
		}
	}

	if i.SplitTypes.Get() {
		opts = append(opts, smfimport.SplitEventTypes())
	}

	if i.Quantize.Get() {
		opts = append(opts, smfimport.Quantize(), smfimport.DetectDelay())
	}

	if i.SkipVelocity.Get() {
		opts = append(opts, smfimport.NoVelocity())
	}

	if i.SkipMicroTiming.Get() {
		opts = append(opts, smfimport.NoMicroTiming())
	}

	return muskel.Import(srcFile, outFile, opts...)
}

func (i *importCmd) watch(a *args) error {
	p, _ := filepath.Abs(i.Source.Get())
	d := p
	if !a.WatchDir.Get() {
		d = filepath.Dir(p)
	}
	targetExt := ".mskl"

	var outFile string

	if i.OutFile.IsSet() {
		outFile, _ = filepath.Abs(i.OutFile.Get())
		targetExt = filepath.Ext(targetExt)
	}

	var callback func(dir, file string) error
	fmt.Printf("watching %q\n", i.Source.Get())

	if a.WatchDir.Get() {
		// TODO get the dir
		callback = func(dir, file string) error {
			ext := strings.ToLower(filepath.Ext(file))

			if ext != ".mid" && ext != ".midi" {
				return nil
			}

			dp, _ := filepath.Abs(dir)

			if dp != p {
				return nil
			}

			filep, _ := filepath.Abs(file)
			err := i.runImport(filep, filep+targetExt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s", err.Error())
			} else {
				fmt.Fprintf(os.Stdout, filep+"\n")
			}
			return err
		}

	} else {

		callback = func(dir, file string) error {
			filep, _ := filepath.Abs(file)

			if filep != p {
				return nil
			}

			if outFile == "" {
				outFile = filep + targetExt
			}

			err := i.runImport(filep, outFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s", err.Error())
			} else {
				fmt.Fprintf(os.Stdout, filep+"\n")
			}
			return err
		}
	}

	r := runfunc.New(
		d,
		callback,
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
*/
