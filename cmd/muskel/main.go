package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/metakeule/config"
)

var (
	cfg = config.MustNew("muskel", "0.9.97", "muskel is a musical sketch language")

	argFile   = cfg.NewString("file", "path of the muskel file", config.Shortflag('f'), config.Required)
	argSketch = cfg.NewString("sketch", "name of the sketch table", config.Shortflag('s'), config.Default("=SCORE"))
	argFlow   = cfg.NewBool("flow", "flow mode; sets sketch to ! and pattern to !", config.Default(false))
	argParams = cfg.NewJSON("params", "parameters passed to the sketch. params must have the syntax [trackname]#[no]:[value] where no is the params number, e.g. [\"voc#2:c#'\",\"piano#1:D\"]", config.Shortflag('p'), config.Default("[]"))
	argPatt   = cfg.NewString("pattern", "pattern to be used exclusively", config.Shortflag('t'), config.Default(""))
	argFmt    = cfg.NewBool("fmt", "format the muskel file (overwrites the input file)")
	//argAddMissing = cfg.NewBool("addprops", "add missing properties")
	argWatch      = cfg.NewBool("watch", "watch for changes of the file and act on each change", config.Shortflag('w'))
	argDir        = cfg.NewBool("dir", "watch for changes in the current directory (not just for the input file)", config.Shortflag('d'))
	argOutFile    = cfg.NewString("out", "path of the output file (SMF). If it includes the placeholder %s, that will be replaced by the File property of the corresponding track", config.Shortflag('o'))
	argTrackFiles = cfg.NewBool("trackfiles", "sets out to '%s' in order to write to the file names as given in the track properties")
	argSleep      = cfg.NewInt32("sleep", "sleeping time between invocations (in milliseconds)", config.Default(int32(100)))
	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
	argUnroll = cfg.NewString("unroll", "unroll the source to the given file name", config.Shortflag('u'))
	argDebug  = cfg.NewBool("debug", "print debug messages")

	cmdSMF      = cfg.MustCommand("smf", "convert a muskel file to Standard MIDI file format (SMF)")
	argSMFTicks = cmdSMF.NewInt32("ticks", "resolution of SMF file in ticks", config.Default(int32(960)))

	cmdImport        = cfg.MustCommand("import", "convert a Standard MIDI file to a muskel file").Relax("file")
	argImportSMF     = cmdImport.NewString("smf", "path of the Standard MIDI file file.", config.Shortflag('i'), config.Required)
	argOutMuskelFile = cmdImport.NewString("mskl", "path of the muskel file.", config.Shortflag('o'), config.Required)

	//cmdAddTrack     = cfg.MustCommand("addtrack", "add a track")
	//argAddTrackName = cmdAddTrack.NewString("name", "name of the track", config.Required)

	//cmdRenameTemplate    = cfg.MustCommand("renametemplate", "rename a template")
	//argRenameTemplateOld = cmdRenameTemplate.NewString("old", "old name of the template", config.Required)
	//argRenameTemplateNew = cmdRenameTemplate.NewString("new", "new name of the template", config.Required)

	//cmdRenameTrack    = cfg.MustCommand("renametrack", "rename a track")
	//argRenameTrackOld = cmdRenameTrack.NewString("old", "old name of the track", config.Required)
	//argRenameTrackNew = cmdRenameTrack.NewString("new", "new name of the track", config.Required)

	//cmdRmTrack     = cfg.MustCommand("rmtrack", "remove a track")
	//argRmTrackName = cmdRmTrack.NewString("name", "name of the track that should be removed", config.Required)

	//cmdSyncTracks = cfg.MustCommand("synctracks", "sync tracks in include files to the tracks within the given main file")

	cmdPlay = cfg.MustCommand("play", "play a muskel file (currently linux only, needs audacious)")
	// timidity $file
	//argPlayCmd = cmdPlay.NewString("cmd", "command to execute when playing", config.Default("audacious -1 -H -p -q $_file"))
	argPlayCmd = cmdPlay.NewString("cmd", "command to execute when playing", config.Default("timidity $_file"))

	cmdTemplate  = cfg.MustCommand("template", "show global template files").Relax("file")
	argTemplFile = cmdTemplate.NewString("template", "file name of the template file", config.Shortflag('t'))
)

func fileCheckSum(file string) string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func fmtFile(file string, params []string, opts ...score.Option) error {
	sc, err := muskel.ParseFile(file, params, opts...)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	/*
		if argSmallCols.Get() {
			sc.SmallColumns = true
		}
	*/

	//sc.AddMissingProperties()

	//err = muskel.WriteFormattedFile(sc, file)
	err = sc.Format()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	if argWatch.Get() {
		fmt.Printf(".")
	}
	return nil
	//beeep.Notify("OK MuSkeL formatted:", "", "assets/information.png")
}

var lock sync.RWMutex

/*
func runAddTrack() error {
	inFile := argFile.Get()
	sketch := argSketch.Get()

	sc, err := muskel.ParseFile(inFile, sketch)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	name := strings.TrimSpace(argAddTrackName.Get())

	if name == "" {
		return fmt.Errorf("empty name is not allowed")
	}
	trck := score.NewTrack(name)

	sc.AddTrack(trck)

	err = muskel.WriteFormattedFile(sc, inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

/*
func runRmTrack() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	name := strings.TrimSpace(argRmTrackName.Get())
	sc.RmTrack(name)

	err = muskel.WriteFormattedFile(sc, inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

/*
func runSyncTracks() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return muskel.SyncTracks(sc)
}
*/

// RenameTrack(old, nu string)
/*
func runRenameTrack() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	old := strings.TrimSpace(argRenameTrackOld.Get())
	nu := strings.TrimSpace(argRenameTrackNew.Get())

	if nu == "" {
		return fmt.Errorf("empty name is not allowed")
	}

	err = muskel.RenameTrack(sc, old, nu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

// RenameTrack(old, nu string)
/*
func runRenamePattern() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	old := strings.TrimSpace(argRenameTemplateOld.Get())
	nu := strings.TrimSpace(argRenameTemplateNew.Get())

	if nu == "" {
		return fmt.Errorf("empty name is not allowed")
	}

	err = muskel.RenameTemplate(sc, old, nu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

type callbackrunner struct {
	inFile  string
	outFile string
	// audacious string
	checksums map[string]string
	ignore    map[string]bool
	inFilep   string
}

func (c *callbackrunner) cmdSMF(sc *score.Score) error {
	err := sc.Unroll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while unrolling MuSkeL score: %s\n", err.Error())
		beeep.Alert("ERROR while unrolling MuSkeL score", err.Error(), "assets/warning.png")
		return err
	}
	err = muskel.WriteSMFFile(sc, c.outFile, smfwriter.TimeFormat(smf.MetricTicks(argSMFTicks.Get())))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
		beeep.Alert("ERROR while converting MuSkeL to SMF", err.Error(), "assets/warning.png")
		return err
	}

	if argWatch.Get() {
		fmt.Fprint(os.Stdout, ".")
	}
	beeep.Notify("OK MuSkeL converted to SMF", path.Base(c.outFile), "assets/information.png")
	return nil
}

var mx sync.Mutex
var playCmd *exec.Cmd
var lastStarted time.Time

func play(cm string) {
	mx.Lock()
	if playCmd != nil {
		playCmd.Process.Kill()
	}
	time.Sleep(time.Second * 1)
	playCmd = execCommand(cm)
	mx.Unlock()
	go func() {
		//cmd := exec.Command("/bin/sh", "e", cm)
		// cmd := exec.Command("audacious", "-1", "-H", "-p", "-q", c.outFile)
		_, _ = playCmd.CombinedOutput()
	}()
}

func (c *callbackrunner) cmdPlay(sc *score.Score) error {
	now := time.Now()
	if now.Sub(lastStarted) < time.Second*2 {
		return nil
	}
	mx.Lock()
	lastStarted = now
	mx.Unlock()

	err := muskel.WriteSMFFile(sc, c.outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
		beeep.Alert("ERROR while converting MuSkeL to SMF", err.Error(), "assets/warning.png")
		return err
	}

	if argWatch.Get() {
		fmt.Fprint(os.Stdout, ".")
	}
	beeep.Notify("OK MuSkeL converted to SMF", path.Base(c.outFile), "assets/information.png")
	cm := strings.TrimSpace(argPlayCmd.Get())
	cm = strings.ReplaceAll(cm, "$_file", c.outFile)
	if cm != "" {
		play(cm)
	}
	//beeep.Alert("audacious not found", err.Error(), "assets/warning.png")
	return fmt.Errorf("play command not defined not found")
}

func writeUnrolled(sc *score.Score) error {
	err := sc.Unroll()

	if err != nil {
		return err
	}

	var uf *os.File

	os.Remove(argUnroll.Get())

	uf, err = os.Create(argUnroll.Get())
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

func (c *callbackrunner) prepare(dir, file string) error {
	//fmt.Printf("prepare(%q, %q) called with %#v\n", dir, file, c)

	if filepath.Ext(file) != ".mskl" {
		return nil
	}

	if strings.Contains(dir, "muskel-fmt") || strings.Contains(file, "muskel-fmt") {
		return nil
	}

	filep, _ := filepath.Abs(file)
	if strings.Contains(filep, "muskel-fmt") {
		return nil
	}

	//fmt.Printf("filep %q vs c.inFilep %q\n", filep, c.inFilep)

	//lock.RLock()
	if c.ignore[filep] {
		//lock.RUnlock()
		return nil
	}

	// fmt.Printf("[%q, %q]\n", dir, file)

	if !argDir.Get() {
		if filep != c.inFilep {
			return nil
		}
	}

	//	lock.Lock()
	newChecksum := fileCheckSum(filep)
	//lock.Unlock()

	if newChecksum == c.checksums[filep] {
		return nil
	}

	//		lock.Lock()
	c.checksums[filep] = newChecksum
	//	lock.Unlock()
	var params []string

	var opts = []score.Option{score.PrintBarComments()}

	if argFlow.Get() {
		//fmt.Println("flow set")
		opts = append(opts, score.Sketch("!"))
		opts = append(opts, score.Column("!"))
	}

	if argSketch.IsSet() && argSketch.Get() != "=SCORE" {
		opts = append(opts, score.Sketch(argSketch.Get()))
	}

	if argDebug.IsSet() {
		opts = append(opts, score.Debug())
	}

	if argPatt.IsSet() && argPatt.Get() != "" {
		opts = append(opts, score.Column(argPatt.Get()))
	}

	if argParams.IsSet() {
		err := argParams.Get(&params)
		if err != nil {
			return fmt.Errorf("invalid params: %s", err.Error())
		}
	}

	if argDir.Get() {
		if filep != c.inFilep {

			//				go func() {
			//fmt.Printf("formatting %q...\n", filep)
			//lock.Lock()
			c.ignore[filep] = true
			//lock.Unlock()

			fmtFile(filep, params, opts...)
			//lock.Lock()
			c.ignore[filep] = false
			//lock.Unlock()
			//fmt.Printf("done formatting %q...\n", filep)
			//			}()

		}
	}

	//lock.Lock()
	//ignore[filep] = true
	defer func() {
		//ignore[filep] = false
		//lock.Unlock()
	}()
	sc, err := muskel.ParseFile(c.inFile, params, opts...)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	/*
		if argSmallCols.Get() {
			sc.SmallColumns = true
		}
	*/

	/*
		if argAddMissing.Get() {
			sc.AddMissingProperties()
		}
	*/

	if argUnroll.IsSet() {

		err = writeUnrolled(sc)
		if err != nil {
			return err
		}

		//fmt.Printf("%v\n", unrolled.Columns)
		// _ = unrolled

		//return smf.New(smffile).Generate(sc)
		//return smf.WriteFile(sc, smffile)
		//var ur *score.Score
		//ur, err = unroller.Unroll(sc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while unrolling MuSkeL: %s\n", err.Error())
			beeep.Alert("ERROR while unrolling MuSkeL:", err.Error(), "assets/warning.png")
			return err
		}

		/*
			err = muskel.WriteFormattedFile(ur, argUnroll.Get(), muskel.CountUnrolledBars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR while writing unrolling MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while writing unrolled MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}
		*/

	}

	if argFmt.Get() {
		//err = muskel.WriteFormattedFile(sc, c.inFilep)
		err = sc.Format()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
			beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
			return err
		}

		//beeep.Notify("OK MuSkeL formatted:", "", "assets/information.png")
	}

	if extIdx := strings.LastIndex(c.outFile, "."); extIdx > 0 && extIdx+1 < len(c.outFile) {
		switch c.outFile[extIdx:] {
		case ".mid", ".midi", ".MID", ".MIDI":
		default:
			c.outFile = c.outFile[:extIdx] + ".mid"
		}
	} else {
		c.outFile = c.outFile + ".mid"
	}

	switch cfg.ActiveCommand() {
	case cmdSMF:
		return c.cmdSMF(sc)
	case cmdPlay:
		return c.cmdPlay(sc)
	default:
		return nil
	}
}

func (c *callbackrunner) mkcallback() (callback func(dir, file string) error) {
	c.checksums = map[string]string{}
	c.ignore = map[string]bool{}
	inFilep, _ := filepath.Abs(c.inFile)
	c.inFilep = inFilep

	callback = func(dir, file string) error {
		return c.prepare(dir, file)
	}
	return
}

func runCmd() (callback func(dir, file string) error, file_, dir_ string) {

	inFile := argFile.Get()
	outFile := inFile
	if argOutFile.IsSet() {
		outFile = argOutFile.Get()
	}

	if argTrackFiles.Get() {
		outFile = "%s"
	}

	/*
		cmd := exec.Command("which", "audacious")
		out, errC := cmd.CombinedOutput()

		audacious := strings.TrimSpace(string(out))

		if errC != nil || audacious == "" {
			audacious = ""
		}
	*/

	dir_ = path.Dir(inFile)
	file_ = path.Base(inFile)

	var cbr callbackrunner
	cbr.inFile = inFile
	cbr.outFile = outFile
	// cbr.audacious = audacious

	callback = cbr.mkcallback()

	return
}

func run() error {

	err := cfg.Run()

	if err != nil {
		fmt.Println(cfg.Usage())
		return err
	}

	if cfg.ActiveCommand() == cmdImport {
		return muskel.Import(argImportSMF.Get(), argOutMuskelFile.Get())
	}

	if cfg.ActiveCommand() == cmdTemplate {
		if argTemplFile.IsSet() && argTemplFile.Get() != "" {
			p := filepath.Join(score.USER_DIR, argTemplFile.Get()+".mskl")
			// p := filepath.Join(score.USERDIR)
			if !score.FileExists(p) {
				return fmt.Errorf("no such template: %q", p)
			}

			fl, err := os.Open(p)

			if err != nil {
				return fmt.Errorf("can't open template %q: %s", p, err.Error())
			}

			defer fl.Close()
			io.Copy(os.Stdout, fl)
			return nil
			//ioutil.ReadFile(muskel.USERDIR)
		} else {
			fls, err := ioutil.ReadDir(score.USER_DIR)
			if err != nil {
				return fmt.Errorf("can't read user directory %q: %s", score.USER_DIR, err.Error())
			}

			if len(fls) == 0 {
				fmt.Fprintf(os.Stdout, "There are no template files stored inside %s\n\n", score.USER_DIR)
				return nil
			}

			fmt.Fprintf(os.Stdout, "The following template files are stored inside %s:\n\n", score.USER_DIR)

			for _, fl := range fls {
				// fmt.Printf("|%-10v|\n", values[i])
				fmt.Fprintf(os.Stdout, "%-24v (modified: %s)\n", fl.Name(), fl.ModTime().Format(time.RFC822))
				// fmt.Fprintf("%10v...\n", values[i])
			}

			return nil
		}
	}

	cmd, file, dir := runCmd()

	lastStarted = time.Now()

	if argWatch.Get() {
		if argDir.Get() {
			fmt.Printf("watching %q\n", dir)
		} else {
			fmt.Printf("watching %q\n", argFile.Get())
		}
		r := runfunc.New(dir, cmd, runfunc.Sleep(time.Millisecond*time.Duration(int(argSleep.Get()))))

		errors := make(chan error, 1)
		stopped, err := r.Run(errors)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		go func() {
			for e := range errors {
				fmt.Fprintf(os.Stderr, "error: %s\n", e)
				_ = e
			}
		}()

		sigchan := make(chan os.Signal, 10)

		// listen for ctrl+c
		go signal.Notify(sigchan, os.Interrupt)

		// interrupt has happend
		<-sigchan
		fmt.Println("\n--interrupted!")

		stopped.Kill()

		os.Exit(0)
		return nil
	}

	/*
		switch cfg.ActiveCommand() {
		case cmdAddTrack:
			return runAddTrack()
		case cmdRmTrack:
			return runRmTrack()
		case cmdSyncTracks:
			return runSyncTracks()
		case cmdRenameTrack:
			return runRenameTrack()
		case cmdRenameTemplate:
			return runRenamePattern()
		}
	*/
	return cmd(dir, filepath.Join(dir, file))

}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

/*
import (
	"fmt"
	"os"

	muskel "gitlab.com/gomidi/muskel_new_new"
)

func main() {
	err := muskel.Convert("main.mskl", "=SCORE", "main.mid")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
*/
