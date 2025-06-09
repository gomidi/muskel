package main

import (
	"fmt"
	"os"

	//"path"
	"path/filepath"
	"strings"

	"gitlab.com/golang-utils/fs/path"

	"bytes"

	"gitlab.com/gomidi/lilypond"
	//	"gitlab.com/gomidi/midi/smf"
	//"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/xlsx"
)

type converter struct {
	inFile    path.Local
	checksums map[string]string
	ignore    map[string]bool
	player    *Player
	dir       path.Local
	file      string
	Config    struct {
		PrintBarComments bool
		Flow             bool
		CutOut           bool
		SoloGroup        uint
		Sketch           string
		Debug            bool
		Pattern          string
		KeepEmptyLines   bool
		Params           []string
		UnrollFile       path.Local
		Fmt              bool
		CSV              string
		XLSX             bool
	}
}

func newConverter(player *Player, a *args) (r *converter) {

	/*
		fsys, err := rootfs.New()

		if err != nil {
			panic(err.Error())
		}
	*/

	inFile := a.InFile.Val

	//	rFile := rootfs.Normalize(inFile)

	c := &converter{
		//fsys:   fsys,
		player: player,
		inFile: inFile,
		dir:    inFile.Dir(),
		file:   path.Base(inFile),
	}

	c.setFromArgs(a)
	return c
}

func (c *converter) setFromArgs(a *args) {
	c.Config.PrintBarComments = true
	c.Config.Flow = a.Flow.Val
	if a.Sketch.IsSet() && a.Sketch.Val != "=SCORE" {
		c.Config.Sketch = a.Sketch.Val
	} else {
		c.Config.Sketch = ""
	}

	c.Config.Debug = a.Debug.Val

	if a.Pattern.IsSet() {
		c.Config.Pattern = a.Pattern.Val
	} else {
		c.Config.Pattern = ""
	}
	c.Config.KeepEmptyLines = a.KeepEmptyLines.Val
	if a.Params.IsSet() {
		c.Config.Params = a.Params.Val
	} else {
		c.Config.Params = nil
	}

	if a.UnrollFile.IsSet() {
		c.Config.UnrollFile = a.UnrollFile.Val
	}

	c.Config.Fmt = a.Fmt.Val

	if muskel.FILE_EXTENSION == ".csv" {
		c.Config.CSV = a.CSVSeparator.Val
		// never fmt when in csv
		c.Config.Fmt = false
	}

	if muskel.FILE_EXTENSION == ".xlsx" {
		c.Config.XLSX = true
		// never fmt when in csv
		c.Config.Fmt = false
	}

	if a.Solo.IsSet() {
		sg := a.Solo.Val
		if sg > 0 {
			c.Config.SoloGroup = uint(sg)
		}
	}

	if a.CutOut.Val {
		c.Config.CutOut = true
	}
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

	if !m.KeepEmptyLines {
		opts = append(opts, score.NoEmptyLines())
	}

	if m.SoloGroup > 0 {
		opts = append(opts, score.SoloGroup(m.SoloGroup))
	}

	if m.CutOut {
		opts = append(opts, score.CutOut())
	}

	if m.CSV != "" {
		opts = append(opts, score.CSV(rune(m.CSV[0])))
	}

	return

}

func (c *converter) cmdSMF(sc *score.Score) error {
	if c.Config.Fmt {
		c.fmtFile(c.inFile, c.Config.Params, c.ScoreOptions()...)
	}

	err := sc.Unroll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while unrolling MuSkeL score: %s\n", err.Error())
		alert("ERROR while unrolling MuSkeL score", err)
		return err
	}

	//	err = muskel.WriteSMFFile(sc, c.player.outFile, smfwriter.TimeFormat(smf.MetricTicks(SMF.ResolutionTicks.Val)))
	err = muskel.WriteSMFFile(sc, c.player.outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
		alert("ERROR while converting MuSkeL to SMF", err)
		return err
	}

	if ARGS.smf.ExportImage.Val {
		err = muskel.WriteImage(sc, path.MustLocal(c.player.outFile.String()+".png"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while exporting to image: %s\n", err.Error())
			alert("ERROR while exporting to image", err)
			return err
		}
	}

	if ARGS.smf.ExportScore.Val {
		err = lilypond.MIDI2PDF(path.ToSystem(c.player.outFile), "", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while exporting score to PDF: %s\n", err.Error())
			alert("ERROR while exporting to score", err)
			return err
		}
	}

	if ARGS.Watch.Val {
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
		var portPlayerStopped = make(chan bool, 10)

		go func() {
			// interrupt has happend
			<-SIGNAL_CHANNEL
			fmt.Println("\n--interrupted!")
			if c.player.playToPort {
				c.player.portOut.Close()
				stopPortPlayer <- true
			}
		}()

		c.player.playOnce(stopPortPlayer, portPlayerStopped)
		if c.player.playToPort {
			<-portPlayerStopped
			c.player.portOut.Close()
		}
	}

	return nil
}

func (r *converter) run() error {
	file := r.dir.Join(r.file)

	if path.IsDir(file) {
		return fmt.Errorf("is directory: %q", file)
	}

	if path.Ext(file) != muskel.FILE_EXTENSION {
		return fmt.Errorf("no muskel file: %q\n", file)
	}

	cb := r.mkcallback()
	return cb(r.dir.String(), r.dir.Join(r.file).String())
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

	if !ARGS.WatchDir.Val {
		if filep != c.inFile.String() {
			return nil
		}
	}

	if ARGS.Watch.Val {
		newChecksum := fileCheckSum(filep)

		if newChecksum == c.checksums[filep] {
			return nil
		}

		c.checksums[filep] = newChecksum
	}

	if ARGS.WatchDir.Val {
		if filep != c.inFile.String() && c.Config.Fmt {
			c.ignore[filep] = true
			c.fmtFile(c.inFile, c.Config.Params, c.ScoreOptions()...)
			c.ignore[filep] = false
		}
	}

	sc, err := c.parseMuskel(c.inFile)

	if err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		alert("ERROR:", err)
		return err
	}

	switch ARGS.App.SelectedCommand() {
	case ARGS.smf.command:
		return c.cmdSMF(sc)
	case ARGS.play.command:
		return c.cmdPlay(sc)
	default:
		return nil
	}
}

func (c *converter) parseMuskel(srcFile path.Local) (*score.Score, error) {
	var opts = c.ScoreOptions()

	sc, err := muskel.ParseFile(srcFile, c.Config.Params, opts...)

	if err != nil {
		return nil, err
	}

	if c.Config.UnrollFile.Relative().String() != "" {
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

func (c *converter) fmtFile(file path.Local, params []string, opts ...score.Option) error {
	if !c.Config.Fmt {
		return nil
	}
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

	if ARGS.Watch.Val {
		fmt.Printf(".")
	}
	return nil
}

func (c *converter) writeUnrolled(file path.Local, sc *score.Score) error {
	err := sc.Unroll()

	if err != nil {
		return err
	}

	var uf *os.File

	os.Remove(file.String())

	uf, err = os.Create(file.String())
	if err != nil {
		return err
	}
	defer uf.Close()
	if filepath.Ext(file.String()) == ".xlsx" {
		var tracksbf, scorebf bytes.Buffer
		err = sc.WriteTracksAndScoreTable(&tracksbf, &scorebf)
		if err != nil {
			return err
		}
		return xlsx.Write(file.String(), tracksbf.String(), scorebf.String())
	} else {
		err = sc.WriteUnrolled(uf)
		if err != nil {
			return err
		}
		return nil
	}
}
