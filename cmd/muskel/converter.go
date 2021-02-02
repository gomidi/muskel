package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/score"
)

type converter struct {
	inFile    string
	checksums map[string]string
	ignore    map[string]bool
	inFilep   string
	player    *Player
	dir       string
	file      string
	Config    struct {
		PrintBarComments bool
		Flow             bool
		Sketch           string
		Debug            bool
		Pattern          string
		KeepEmptyLines   bool
		Params           []string
		UnrollFile       string
		Fmt              bool
	}
}

func newConverter(player *Player, a *args) (r *converter) {
	inFile := a.File.Get()

	c := &converter{
		player: player,
		inFile: inFile,
		dir:    path.Dir(inFile),
		file:   path.Base(inFile),
	}
	inFilep, _ := filepath.Abs(c.inFile)
	c.inFilep = inFilep

	c.setFromArgs(a)
	return c
}

func (c *converter) setFromArgs(a *args) {
	c.Config.PrintBarComments = true
	c.Config.Flow = a.Flow.Get()
	if a.Sketch.IsSet() && a.Sketch.Get() != "=SCORE" {
		c.Config.Sketch = a.Sketch.Get()
	} else {
		c.Config.Sketch = ""
	}

	c.Config.Debug = a.Debug.Get()

	if a.Pattern.IsSet() {
		c.Config.Pattern = a.Pattern.Get()
	} else {
		c.Config.Pattern = ""
	}
	c.Config.KeepEmptyLines = a.KeepEmptyLines.Get()
	if a.Params.IsSet() {
		c.Config.Params = []string{}
		a.Params.Get(&c.Config.Params)
	} else {
		c.Config.Params = nil
	}

	if a.UnrollFile.IsSet() {
		c.Config.UnrollFile = a.UnrollFile.Get()
	}

	c.Config.Fmt = a.Fmt.Get()
}

func (c *converter) ScoreOptions() (opts []score.Option) {
	m := c.Config
	if m.PrintBarComments {
		opts = append(opts, score.PrintBarComments())
	}

	if m.Flow {
		opts = append(opts, score.Sketch("!"))
		opts = append(opts, score.Column("!"))
	}

	if m.Sketch != "" {
		opts = append(opts, score.Sketch(m.Sketch))
	}

	if m.Debug {
		opts = append(opts, score.Debug())
	}

	if m.Pattern != "" {
		opts = append(opts, score.Column(m.Pattern))
	}

	if m.KeepEmptyLines {
		opts = append(opts, score.NoEmptyLines())
	}

	return

}

func (c *converter) cmdSMF(sc *score.Score) error {
	if ARGS.Fmt.Get() {
		c.fmtFile(c.inFile, c.Config.Params, c.ScoreOptions()...)
	}

	err := sc.Unroll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while unrolling MuSkeL score: %s\n", err.Error())
		alert("ERROR while unrolling MuSkeL score", err)
		return err
	}

	err = muskel.WriteSMFFile(sc, c.player.outFile, smfwriter.TimeFormat(smf.MetricTicks(SMF.ResolutionTicks.Get())))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
		alert("ERROR while converting MuSkeL to SMF", err)
		return err
	}

	if ARGS.Watch.Get() {
		fmt.Fprint(os.Stdout, ".")
	}
	notify("OK MuSkeL converted to SMF", path.Base(c.player.outFile))
	return nil
}

func (c *converter) cmdPlay(sc *score.Score) error {
	err := c.cmdSMF(sc)
	if err != nil {
		fmt.Println("error while creating smf")
		return err
	}

	if c.player.playerEnabled {
		var stopPortPlayer = make(chan bool, 1)
		var portPlayerStopped = make(chan bool, 1)

		go func() {
			// interrupt has happend
			<-SIGNAL_CHANNEL
			fmt.Println("\n--interrupted!")
			if c.player.playToPort {
				stopPortPlayer <- true
			}
		}()

		c.player.playOnce(stopPortPlayer, portPlayerStopped)
		if c.player.playToPort {
			<-portPlayerStopped
		}
	}

	return nil
}

func (r *converter) run() error {
	return r.mkcallback()(r.dir, filepath.Join(r.dir, r.file))
}

func (c *converter) mkcallback() (callback func(dir, file string) error) {
	c.checksums = map[string]string{}
	c.ignore = map[string]bool{}

	callback = func(dir, file string) error {
		return c.prepare(dir, file)
	}
	return
}

func (c *converter) prepare(dir, file string) error {
	if filepath.Ext(file) != muskel.FILE_EXTENSION {
		return nil
	}

	if strings.Contains(dir, "muskel-fmt") || strings.Contains(file, "muskel-fmt") {
		return nil
	}

	filep, _ := filepath.Abs(file)
	if strings.Contains(filep, "muskel-fmt") {
		return nil
	}

	if c.ignore[filep] {
		return nil
	}

	if !ARGS.WatchDir.Get() {
		if filep != c.inFilep {
			return nil
		}
	}

	newChecksum := fileCheckSum(filep)

	if newChecksum == c.checksums[filep] {
		return nil
	}

	c.checksums[filep] = newChecksum

	if ARGS.WatchDir.Get() {
		if filep != c.inFilep {
			c.ignore[filep] = true
			c.fmtFile(filep, c.Config.Params, c.ScoreOptions()...)
			c.ignore[filep] = false
		}
	}

	sc, err := c.parseMuskel(c.inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		alert("ERROR:", err)
		return err
	}

	switch CONFIG.ActiveCommand() {
	case SMF.Config:
		return c.cmdSMF(sc)
	case PLAY.Config:
		return c.cmdPlay(sc)
	default:
		return nil
	}
}

func (c *converter) parseMuskel(srcFile string) (*score.Score, error) {
	var opts = c.ScoreOptions()

	sc, err := muskel.ParseFile(srcFile, c.Config.Params, opts...)

	if err != nil {
		return nil, fmt.Errorf("ERROR while parsing MuSkeL: %s", err.Error())
	}

	if c.Config.UnrollFile != "" {
		err = c.writeUnrolled(c.Config.UnrollFile, sc)
		if err != nil {
			return sc, fmt.Errorf("ERROR while unrolling MuSkeL: %s", err.Error())
		}
	}

	if c.Config.Fmt {
		err = sc.Format()
		if err != nil {
			return sc, fmt.Errorf("ERROR while writing formatting MuSkeL: %s", err.Error())
		}
	}

	return sc, err

}

func (c *converter) fmtFile(file string, params []string, opts ...score.Option) error {
	sc, err := muskel.ParseFile(file, params, opts...)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		alert("ERROR while parsing MuSkeL:", err)
		return err
	}

	err = sc.Format()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		alert("ERROR while writing formatting MuSkeL:", err)
		return err
	}

	if ARGS.Watch.Get() {
		fmt.Printf(".")
	}
	return nil
}

func (c *converter) writeUnrolled(file string, sc *score.Score) error {
	err := sc.Unroll()

	if err != nil {
		return err
	}

	var uf *os.File

	os.Remove(file)

	uf, err = os.Create(file)
	if err != nil {
		return err
	}
	defer uf.Close()
	err = sc.WriteUnrolled(uf)
	if err != nil {
		return err
	}
	return nil
}
