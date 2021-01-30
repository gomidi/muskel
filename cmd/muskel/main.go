package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	//"github.com/gen2brain/beeep"
	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/metakeule/config"
)

var (
	cfg = config.MustNew("muskel", muskel.VERSION, "muskel is a musical sketch language")

	argFile                = cfg.LastString("file", "path of the muskel file", config.Required)
	argIgnoreMuskelVersion = cfg.NewBool("current", "use the current version of the muskel command and ignore the "+muskel.MUSKEL_VERSION_FILE+" file", config.Default(false), config.Shortflag('c'))
	argSketch              = cfg.NewString("sketch", "name of the sketch table", config.Shortflag('s'), config.Default("=SCORE"))
	argFlow                = cfg.NewBool("flow", "flow mode; sets sketch to ! and pattern to !", config.Default(false))
	argParams              = cfg.NewJSON("params", "parameters passed to the sketch. params must have the syntax [trackname]#[no]:[value] where no is the params number, e.g. [\"voc#2:c#'\",\"piano#1:D\"]", config.Shortflag('p'), config.Default("[]"))
	argPatt                = cfg.NewString("pattern", "pattern to be used exclusively", config.Shortflag('t'), config.Default(""))
	argFmt                 = cfg.NewBool("fmt", "format the muskel file (overwrites the input file)")
	argNoEmptyLines        = cfg.NewBool("noemptylines", "don't remove empty lines from the score", config.Shortflag('e'))
	argWatch               = cfg.NewBool("watch", "watch for changes of the file and act on each change", config.Shortflag('w'))
	argDir                 = cfg.NewBool("dir", "watch for changes in the current directory (not just for the input file)", config.Shortflag('d'))
	argOutFile             = cfg.NewString("out", "path of the output file (SMF). If it includes the placeholder %s, that will be replaced by the File property of the corresponding track", config.Shortflag('o'))
	argTrackFiles          = cfg.NewBool("trackfiles", "sets out to '%s' in order to write to the file names as given in the track properties")
	argSleep               = cfg.NewInt32("sleep", "sleeping time between invocations (in milliseconds)", config.Default(int32(100)))
	argPlayServerAddr      = cfg.NewString("addr", "address of the command/play server", config.Default("localhost:8800"))

	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
	argUnroll = cfg.NewString("unroll", "unroll the source to the given file name", config.Shortflag('u'))
	argDebug  = cfg.NewBool("debug", "print debug messages")

	cmdSMF      = cfg.MustCommand("smf", "convert a muskel file to Standard MIDI file format (SMF)")
	argSMFTicks = cmdSMF.NewInt32("ticks", "resolution of SMF file in ticks", config.Default(int32(960)))

	cmdImport        = cfg.MustCommand("import", "convert a Standard MIDI file to a muskel file").Skip("file").Skip("sketch").Skip("flow").Skip("params").Skip("pattern").Skip("noemptylines").Skip("trackfiles").Skip("unroll").Skip("debug").Skip("fmt").Skip("dir").Skip("out").Skip("watch").Skip("sleep").Skip("current")
	argImportSMF     = cmdImport.LastString("midifile", "the source SMF/MIDI file.", config.Required)
	argOutMuskelFile = cmdImport.NewString("out", "the target muskel file.", config.Shortflag('o'), config.Required)
	argMonoTracks    = cmdImport.NewString("mono", "tracks that are considered to be monophon (e.g. 0,4,5)", config.Shortflag('m'))
	argDrumTracks    = cmdImport.NewString("onset", "tracks that only have onset data (e.g. drum notes of or keyswitches) (e.g. 0,4,5)", config.Shortflag('r'))

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

	cmdPlay = cfg.MustCommand("play", "play a muskel file (currently linux only, needs timidity)")
	//argPlayCmd = cmdPlay.NewString("cmd", "command to execute when playing", config.Default("audacious -1 -H -p -q $_file"))
	argPlayCmd    = cmdPlay.NewString("cmd", "command to execute when playing (fluidsynth,timidity,audacious or custom (pass $_file variable))", config.Default("fluidsynth"))
	argPlayServer = cmdPlay.NewBool("server", "start command server for playing and stopping")
	cmdTemplate   = cfg.MustCommand("template", "show global template files").SkipAllBut()
	argTemplFile  = cmdTemplate.LastString("templatefile", "file name of the template file that should be shown. If no file is given, the list of available files is shown.")

	cmdPlayClient   = cfg.MustCommand("client", "client to the play server. allows play and stop").SkipAllBut("addr")
	argClientToggle = cmdPlayClient.NewBool("toggle", "toggle between play/stop playing the current file", config.Shortflag('t'))
	argClientPlay   = cmdPlayClient.NewBool("play", "play (true) or stop (false) playing the current file", config.Shortflag('p'))

	cmdConfigDirs = cfg.MustCommand("dirs", "show configuration dirs").SkipAllBut()
)

var cmdMaps = map[string]string{
	"timidity":   "timidity --quiet -V linear --noise-shaping=1 $_file",
	"audacious":  "audacious -1 -H -p -q $_file",
	"fluidsynth": defaultPlayCmd(),
}

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
		alert("ERROR while parsing MuSkeL:", err)
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
		alert("ERROR while writing formatting MuSkeL:", err)
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

func serverPlay(wr http.ResponseWriter, r *http.Request) {
	playCh <- true
	wr.Write([]byte("start playing"))
}

func serverStop(wr http.ResponseWriter, r *http.Request) {
	stopCh <- true
	<-stoppedCh
	wr.Write([]byte("stopped playing"))
}

func serverToggle(wr http.ResponseWriter, r *http.Request) {
	toggleCh <- true
	wr.Write([]byte("toggle playing"))
}

func runServer() {
	if !argPlayServer.Get() {
		return
	}

	addr := argPlayServerAddr.Get()

	http.HandleFunc("/play", serverPlay)
	http.HandleFunc("/stop", serverStop)
	http.HandleFunc("/toggle", serverToggle)

	/*
		http.Handle("/play", http.HandleFunc(r *http.Request, wr http.ResponseWriter) {

		})
	*/

	go http.ListenAndServe(addr, nil)
}

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
		alert("ERROR while unrolling MuSkeL score", err)
		return err
	}
	err = muskel.WriteSMFFile(sc, c.outFile, smfwriter.TimeFormat(smf.MetricTicks(argSMFTicks.Get())))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
		alert("ERROR while converting MuSkeL to SMF", err)
		return err
	}

	if argWatch.Get() {
		fmt.Fprint(os.Stdout, ".")
	}
	notify("OK MuSkeL converted to SMF", path.Base(c.outFile))
	return nil
}

var mx sync.Mutex

//var playCmd *exec.Cmd
//var lastStarted time.Time
//var playWG sync.WaitGroup
var playCh = make(chan bool, 1)

//var playerPlay = make(chan bool, 1)
var stopCh = make(chan bool, 1)
var stopPlayer = make(chan bool, 1)
var stoppedCh = make(chan bool, 1)
var finishedCh = make(chan bool, 1)
var playerStopped = make(chan bool, 1)
var gotPid = make(chan int, 1)
var toggleCh = make(chan bool, 1)

var playCmdString string

func player() {
	var cmd *exec.Cmd
	var pid int
	//var cm string
	//var wg sync.WaitGroup

	for {
		select {
		case <-toggleCh:
			if cmd != nil {
				killCmd(cmd, pid)
				cmd = nil
			} else {
				cmd = execCommand(playCmdString)
				//wg.Add(1)
				go func() {
					pid_ := runThePlayerCommand(playCmdString, cmd)
					gotPid <- pid_
					//wg.Done()
					//finishedCh <- true
				}()
			}
		case <-stopPlayer:
			if cmd != nil {
				killCmd(cmd, pid)
				cmd = nil
			}
			playerStopped <- true
			return
		case pid = <-gotPid:
			//fmt.Printf("got pid %v\n", pid)
		case <-playCh:
			//fmt.Println("received play")
			if cmd != nil {
				killCmd(cmd, pid)
				cmd = nil
			}
			//fmt.Println("everything stopped")
			cmd = execCommand(playCmdString)
			//wg.Add(1)
			go func() {
				pid_ := runThePlayerCommand(playCmdString, cmd)
				gotPid <- pid_
				//wg.Done()
				//finishedCh <- true
			}()
		case <-stopCh:
			if cmd != nil {
				killCmd(cmd, pid)
				cmd = nil
			}
			stoppedCh <- true
			/*
				case <-finishedCh:
					fmt.Println("finished")
					cmd = nil
			*/
		default:
		}
	}
}

func runThePlayerCommand(cm string, cmd *exec.Cmd) int {
	var err error
	if argWatch.Get() {
		err = cmd.Start()
		return cmd.Process.Pid
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", cm)
		var b []byte
		b, err = cmd.CombinedOutput()
		if len(b) > 0 {
			os.Stdout.Write(b)
		}
	}
	/*
		b, err := cmd.CombinedOutput()
		if len(b) > 0 {
			os.Stdout.Write(b)
		}
	*/
	//
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while running %q: %v", cm, err)
	}

	return -1
}

func play() {
	//func play() {
	if argWatch.Get() {
		if !argPlayServer.Get() {
			playCh <- true
		}
		return
	}
	//cmd := exec.Command("/bin/sh", "e", cm)
	// cmd := exec.Command("audacious", "-1", "-H", "-p", "-q", c.outFile)
	runThePlayerCommand(playCmdString, execCommand(playCmdString))
}

func (c *callbackrunner) mkPlayCmdString() string {
	cm := strings.TrimSpace(argPlayCmd.Get())

	if cc, has := cmdMaps[cm]; has {
		cm = cc
	}

	cm = strings.ReplaceAll(cm, "$_file", c.outFile)
	return cm
}

func (c *callbackrunner) cmdPlay(sc *score.Score) error {
	//fmt.Println("running play")
	err := c.cmdSMF(sc)
	if err != nil {
		fmt.Println("error while creating smf")
		return err
	}

	/*
		if argWatch.Get() {
			fmt.Println(time.Now().Sub(lastStarted))
			now := time.Now()
			if now.Sub(lastStarted) < 300*time.Millisecond {
				return nil
			}
			mx.Lock()
			lastStarted = now
			mx.Unlock()
		}
	*/

	/*
		cm := strings.TrimSpace(argPlayCmd.Get())
		cm = strings.ReplaceAll(cm, "$_file", c.outFile)
		if cm != "" {
			//fmt.Printf("playing")
			play(cm)
		}
	*/
	playCmdString = c.mkPlayCmdString()
	play()

	//beeep.Alert("audacious not found", err.Error(), "assets/warning.png")
	//return fmt.Errorf("play command not defined not found")
	return nil
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
	//fmt.Printf("prepare(%q, %q) called with %#v file extension: %q\n", dir, file, c, muskel.FILE_EXTENSION)

	if filepath.Ext(file) != muskel.FILE_EXTENSION {
		//return fmt.Errorf("wrong file extension %q (expected %q)", filepath.Ext(file), muskel.FILE_EXTENSION)
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

	if argNoEmptyLines.Get() {
		opts = append(opts, score.NoEmptyLines())
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
		alert("ERROR while parsing MuSkeL:", err)
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
			alert("ERROR while unrolling MuSkeL:", err)
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
			alert("ERROR while writing formatting MuSkeL:", err)
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

func configDirs() error {
	fmt.Fprintf(os.Stdout, "USER_DIR: %q\nWORKING_DIR: %q\n\n", muskel.USER_DIR, muskel.WORKING_DIR)
	return nil
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

func runCmd() (cbru *callbackrunner, file_, dir_ string) {

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

	//callback = cbr.mkcallback()
	cbru = &cbr

	return
}

func run() error {

	err := cfg.Run()

	if err != nil {
		fmt.Println(cfg.Usage())
		return err
	}

	//fmt.Println("hello")

	if cfg.ActiveCommand() == cmdConfigDirs {
		return configDirs()
	}

	if cfg.ActiveCommand() == cmdPlayClient {
		var r *http.Response
		var err error
		addr := "http://" + argPlayServerAddr.Get() + "/"
		if argClientToggle.Get() {
			r, err = http.Get(addr + "toggle")
		} else {
			if argClientPlay.IsSet() {
				if argClientPlay.Get() {
					r, err = http.Get(addr + "play")
				} else {
					r, err = http.Get(addr + "stop")
				}
			}
		}

		if r != nil {
			ioutil.ReadAll(r.Body)
			r.Body.Close()
		}

		return err
	}

	if cfg.ActiveCommand() == cmdImport {
		var monoTracks []int
		var drumTracks []int

		if argMonoTracks.IsSet() {
			ms := strings.Split(argMonoTracks.Get(), ",")

			for _, m := range ms {
				i, errI := strconv.Atoi(strings.TrimSpace(m))
				if errI == nil {
					monoTracks = append(monoTracks, i)
				}
			}
		}

		if argDrumTracks.IsSet() {
			ms := strings.Split(argDrumTracks.Get(), ",")

			for _, m := range ms {
				i, errI := strconv.Atoi(strings.TrimSpace(m))
				if errI == nil {
					drumTracks = append(drumTracks, i)
				}
			}
		}

		return muskel.Import(argImportSMF.Get(), argOutMuskelFile.Get(), monoTracks, drumTracks)
	}

	if cfg.ActiveCommand() == cmdTemplate {
		if argTemplFile.IsSet() && argTemplFile.Get() != "" {
			p := filepath.Join(muskel.USER_DIR, argTemplFile.Get()+muskel.FILE_EXTENSION)
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
			fls, err := ioutil.ReadDir(muskel.USER_DIR)
			if err != nil {
				return fmt.Errorf("can't read user directory %q: %s", muskel.USER_DIR, err.Error())
			}

			if len(fls) == 0 {
				fmt.Fprintf(os.Stdout, "There are no template files stored inside %s\n\n", muskel.USER_DIR)
				return nil
			}

			fmt.Fprintf(os.Stdout, "The following template files are stored inside %s:\n\n", muskel.USER_DIR)

			for _, fl := range fls {
				// fmt.Printf("|%-10v|\n", values[i])
				fmt.Fprintf(os.Stdout, "%-24v (modified: %s)\n", fl.Name(), fl.ModTime().Format(time.RFC822))
				// fmt.Fprintf("%10v...\n", values[i])
			}

			return nil
		}
	}

	if filepath.Ext(argFile.Get()) == ".md" {
		//fmt.Println("setting file extension to .md")
		muskel.FILE_EXTENSION = ".md"
	}

	srcdir := filepath.Dir(argFile.Get())

	if !argIgnoreMuskelVersion.Get() {
		var v *config.Version
		v, err = muskel.ReadWDVersionFile(srcdir)
		if err == nil {
			if v.String() != muskel.VERSION {
				name := versionate("muskel", v)
				fmt.Fprintf(os.Stderr, "This is version "+muskel.VERSION+" of muskel and your "+muskel.MUSKEL_VERSION_FILE+
					" points to version "+v.String()+
					".\nTo preserve compatibility with your musical notation, the binary \""+name+"\" will be called.\nIf you don't want this behavior or have no such older versioned muskel file, "+
					"\nremove the file \""+filepath.Join(srcdir, muskel.MUSKEL_VERSION_FILE)+"\"\nor pass the \"--current\" option to let your file be parsed anyway.\n\n")

				cmd := runVersionated(name, os.Args[1:])
				cmd.Dir, _ = os.Getwd()
				cmd.Env = os.Environ()
				var out []byte
				out, err = cmd.CombinedOutput()
				if err != nil {
					if _, notExeced := err.(*exec.Error); notExeced {
						fmt.Fprintf(os.Stderr, "Error: could not find %q binary\n", name)
						os.Exit(1)
					}
					fmt.Fprintln(os.Stderr, string(out))
					os.Exit(1)
				}
				fmt.Fprintln(os.Stdout, string(out))
				os.Exit(0)
			}
		}
		muskel.WriteWDVersionFile(srcdir)
	}

	cbr, file, dir := runCmd()

	//if argWatch.Get() {
	//lastStarted = time.Now()
	//time.Sleep(800 * time.Millisecond)
	//}
	//lastStarted = time.Now()

	if argWatch.Get() {
		if argDir.Get() {
			fmt.Printf("watching %q\n", dir)
		} else {
			fmt.Printf("watching %q\n", argFile.Get())
		}

		if cfg.ActiveCommand() == cmdPlay {
			//fmt.Println("starting player loop")
			go player()
		}

		runServer()

		r := runfunc.New(dir, cbr.mkcallback(), runfunc.Sleep(time.Millisecond*time.Duration(int(argSleep.Get()))))

		errors := make(chan error, 1)
		stopped, err := r.Run(errors)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		go func() {
			for {
				select {
				case er := <-errors:
					fmt.Fprintf(os.Stderr, "error: %s\n", er)
				default:
				}
			}
			/*
				for e := range errors {
					fmt.Fprintf(os.Stderr, "error: %s\n", e)
					_ = e
				}
			*/
		}()

		sigchan := make(chan os.Signal, 10)

		// listen for ctrl+c
		go signal.Notify(sigchan, os.Interrupt)

		// interrupt has happend
		<-sigchan
		fmt.Println("\n--interrupted!")
		if cfg.ActiveCommand() == cmdPlay {
			stopPlayer <- true
			<-playerStopped
		}

		stopped.Kill()

		os.Exit(0)
		return nil
	}

	return cbr.mkcallback()(dir, filepath.Join(dir, file))

}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}
