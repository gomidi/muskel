package converter

import (
	"bytes"
	"fmt"
	"os"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/config"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/oshelpers"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/player"
	"gitlab.com/gomidi/muskel/lilypond"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/xlsx"
	"gitlab.com/gomidi/smfplayer"
)

type Converter struct {
	Player      *player.Player
	InFile      path.Local
	Dir         path.Local
	File        string
	Checksums   map[string]string
	Ignore      map[string]bool
	Config      config.Config
	doWatch     bool
	SmfAudioout *smfplayer.AudioOut
}

func New(pl *player.Player, inFile path.Local, doWatch bool, smfAudioout *smfplayer.AudioOut) *Converter {
	if pl != nil && pl.Enabled && !pl.RenderToWave && !pl.PlayToPort && smfAudioout == nil {
		panic("audioout must be passed when playing an smf file with smfplayer")
	}
	return &Converter{
		Player:      pl,
		SmfAudioout: smfAudioout,
		InFile:      inFile,
		doWatch:     doWatch,
		Dir:         inFile.Dir(),
		File:        path.Base(inFile),
		Checksums:   map[string]string{},
		Ignore:      map[string]bool{},
	}
}

func (me *Converter) ScoreOptions() (opts []score.Option) {
	m := me.Config
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

func (me *Converter) CmdSMF(sc *score.Score) error {

	if me.Config.Fmt {
		me.FmtFile(me.InFile, me.Config.Params, me.ScoreOptions()...)
	}

	err := sc.Unroll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while unrolling MuSkeL score: %s\n", err.Error())
		oshelpers.Alert("ERROR while unrolling MuSkeL score", err)
		return err
	}

	//	err = muskel.WriteSMFFile(sc, c.player.outFile, smfwriter.TimeFormat(smf.MetricTicks(SMF.ResolutionTicks.Val)))
	err = muskel.WriteSMFFile(sc, me.Player.SMFFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
		oshelpers.Alert("ERROR while converting MuSkeL to SMF", err)
		return err
	}

	var locImgFile path.Local

	if me.Config.SMF.ExportImage {
		locImgFile = path.MustLocal(me.Player.SMFFile.String() + ".png")
		err = muskel.WriteImage(sc, locImgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while exporting to image: %s\n", err.Error())
			oshelpers.Alert("ERROR while exporting to image", err)
			return err
		}
	}

	if me.Config.SMF.ExportScore {
		err = lilypond.WriteFile(sc, me.Config.SMF.PDFfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while exporting score to PDF: %s\n", err.Error())
			oshelpers.Alert("ERROR while exporting to score", err)
			return err
		}
	}

	if me.Config.SMF.Watch {
		fmt.Fprint(os.Stdout, ".")
	}

	oshelpers.Notify("OK MuSkeL converted to SMF", path.Base(me.Player.SMFFile))

	if me.Config.SMF.Open {

		// open possibly all exported files
		if me.Config.SMF.Watch {
			//open(c.player.outFile, true)

			if me.Config.SMF.ExportImage {
				oshelpers.Open(locImgFile, true)
			}

			if me.Config.SMF.ExportScore {
				oshelpers.Open(me.Config.SMF.PDFfile, true)
			}
			me.Player.ConvertAndPlay(me.SmfAudioout)
			return nil
		}

		// just open one of them
		switch {
		case me.Config.SMF.ExportScore:
			// open the score pdf
			oshelpers.Open(me.Config.SMF.PDFfile, false)
		case me.Config.SMF.ExportImage:
			// open the image
			oshelpers.Open(locImgFile, false)
		default:
			// open the smf
			me.Player.ConvertAndPlay(me.SmfAudioout)
			//open(c.player.outFile, false)
		}
	}

	return nil
}

func (me *Converter) CmdPlay(signl_ch chan os.Signal, sc *score.Score, watching bool) error {
	err := me.CmdSMF(sc)
	if err != nil {
		fmt.Println("error while creating smf")
		return err
	}

	shouldClose := !watching

	if me.Player.Enabled {
		var stopPortPlayer = make(chan bool, 1)
		var portPlayerStopped = make(chan bool, 10)

		go func() {
			// interrupt has happend
			<-signl_ch
			fmt.Println("\n--interrupted (CTRL+C)!")
			/*
				if me.Player.PlayToPort {
					fmt.Println("closing port (converter)")
					me.Player.PortOut.Close()
				}
			*/
			//	fmt.Println("sending stopPortPlayer")
			// trigger the stopping
			shouldClose = true
			stopPortPlayer <- true
		}()

		me.Player.PlayOnce(stopPortPlayer, portPlayerStopped, me.SmfAudioout)
		//	fmt.Println("waiting for portPlayerStopped")
		// wait until the stopping finished
		<-portPlayerStopped

		if me.Player.PlayToPort && shouldClose {
			me.Player.PortOut.Close()
		}
	}

	return nil
}

func (me *Converter) ParseMuskel(srcFile path.Local) (*score.Score, error) {
	var opts = me.ScoreOptions()

	sc, err := muskel.ParseFile(srcFile, me.Config.Params, opts...)

	if err != nil {
		return nil, err
	}

	if me.Config.UnrollFile != "" {
		err = me.writeUnrolled(me.Config.UnrollFile, sc)
		if err != nil {
			return sc, fmt.Errorf("ERROR while unrolling MuSkeL: %s", err.Error())
		}
	}

	if me.Config.Fmt {
		err = sc.Format()
		if err != nil {
			return sc, fmt.Errorf("ERROR while writing formatting MuSkeL: %s", err.Error())
		}
	}

	return sc, err

}

func (me *Converter) FmtFile(file path.Local, params []string, opts ...score.Option) error {
	if !me.Config.Fmt {
		return nil
	}
	sc, err := muskel.ParseFile(file, params, opts...)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		oshelpers.Alert("ERROR while parsing MuSkeL:", err)
		return err
	}

	err = sc.Format()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		oshelpers.Alert("ERROR while writing formatting MuSkeL:", err)
		return err
	}

	if me.doWatch {
		fmt.Printf(".")
	}
	return nil
}

func (me *Converter) writeUnrolled(file string, sc *score.Score) error {
	err := sc.Unroll()

	if err != nil {
		return err
	}

	fileRel := score.ProjectPath(file).Relative(sc.ProjectDir)

	sc.FS.Delete(fileRel, false)
	uf, err := fs.OpenWriter(sc.FS, fileRel)

	if err != nil {
		return err
	}
	defer uf.Close()
	if path.Ext(fileRel) == ".xlsx" {
		var tracksbf, scorebf bytes.Buffer
		err = sc.WriteTracksAndScoreTable(&tracksbf, &scorebf)
		if err != nil {
			return err
		}
		return xlsx.Write(path.ToSystem(sc.FS.Abs(fileRel)), tracksbf.String(), scorebf.String())
	} else {
		err = sc.WriteUnrolled(uf)
		if err != nil {
			return err
		}
		return nil
	}
}
