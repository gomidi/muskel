package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/golang-utils/config/v3"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/fsutils/fsshell"
	"gitlab.com/golang-utils/version/v2"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/converter"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/oshelpers"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/player"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/server"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/smfplayer"
)

var (
	//	APP            = &app{}
	SIGNAL_CHANNEL = make(chan os.Signal, 10)
)

func New() *App {
	return (&App{}).init()
}

type App struct {
	*config.App
	InFile         config.FileLocal
	CSVSeparator   config.String
	Current        config.Bool
	Sketch         config.String
	Flow           config.Bool
	Solo           config.Int
	CutOut         config.Bool
	Params         config.JSON[[]string]
	Pattern        config.String
	Fmt            config.Bool
	KeepEmptyLines config.Bool
	Watch          config.Bool
	WatchDir       config.Bool
	OutFile        config.FileLocal
	TrackFiles     config.Bool
	SleepingTime   config.Int
	UnrollFile     config.String
	Debug          config.Bool

	smf struct {
		command         *config.Command
		ResolutionTicks config.Int
		ExportImage     config.Bool
		ExportScore     config.Bool
		PDFfile         config.FileLocal
		Open            config.Bool
	}

	ports struct {
	}

	smfJSON struct {
		File    config.FileLocal
		OutFile config.FileLocal
	}

	template struct {
		File config.String
	}

	play struct {
		command *config.Command
		//Program config.Select[string] // config.String // fluidsynth,timidity,audacious,auto
		OutPort  config.Int
		WaveFile config.FileLocal
	}

	server struct {
		command      *config.Command
		Address      config.String
		PlayProgram  config.String
		playershared *player.Player
	}

	client struct {
		ServerAddress config.String
		Toggle        config.Bool
		Play          config.Bool
		Convert       config.Bool
	}

	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
}

func (me *App) newPlayer() (p *player.Player, err error) {
	p = player.New()
	err = me.setupPlayer(p)
	return
}

func (me *App) normalizePlayerOutFile(pl *player.Player) {
	pl.SMFFile = me.InFile.Val
	if me.OutFile.IsSet() {
		pl.SMFFile = me.OutFile.Val
	}

	if me.TrackFiles.Val {
		pl.SMFFile = me.OutFile.Val.Dir().Join("%s")
	}

	if extIdx := strings.LastIndex(pl.SMFFile.String(), "."); extIdx > 0 && extIdx+1 < len(pl.SMFFile.String()) {
		switch pl.SMFFile.String()[extIdx:] {
		case ".mid", ".midi", ".MID", ".MIDI":
		default:
			pl.SMFFile = path.MustLocal(pl.SMFFile.String()[:extIdx] + ".mid")
		}
	} else {
		pl.SMFFile = path.MustLocal(pl.SMFFile.String() + ".mid")
	}
}

func (me *App) setupPlayer(pl *player.Player) error {
	me.normalizePlayerOutFile(pl)

	if me.App.SelectedCommand() != me.play.command && me.App.SelectedCommand() != me.server.command {
		return nil
	}

	pl.Enabled = true
	pl.PlayToPort = false

	switch {
	case me.play.OutPort.IsSet():
		port := me.play.OutPort.Val
		fmt.Printf("out port: %v\n", port)
		/*
			drv, err := newDriver()
			if err != nil {
				return fmt.Errorf("can't open midi driver")
				}
		*/
		outs := midi.GetOutPorts()
		if len(outs)-1 < port {
			return fmt.Errorf("can't open midi out port %v", port)
		}
		//ps.portOut, err = midi.OpenOut(drv, p, "")
		pl.PlayToPort = true
		pl.PortOut = outs[port]
		//if err != nil {
		//	return fmt.Errorf("can't open midi out port %v", me.play.OutPort.Val)
		//}
	case me.play.WaveFile.IsSet():
		pl.WavFile = me.play.WaveFile.Val
		pl.RenderToWave = true
	default:
	}

	return nil
}

func (me *App) runClient() error {
	var r *http.Response
	var err error
	addr := "http://" + me.client.ServerAddress.Val + "/"
	switch {
	case me.client.Play.IsSet():
		if me.client.Play.Val {
			r, err = http.Get(addr + "play")
		} else {
			r, err = http.Get(addr + "stop")
		}
	case me.client.Convert.IsSet():
		if me.client.Convert.Val {
			r, err = http.Get(addr + "convert")
		}
	case me.client.Toggle.IsSet():
		if me.client.Toggle.Val {
			r, err = http.Get(addr + "toggle")
		}
	case me.Solo.IsSet():
		r, err = http.Get(addr + fmt.Sprintf("solo?solo=%v", me.Solo.Val))
	case me.CutOut.IsSet():
		if me.CutOut.Val {
			r, err = http.Get(addr + "cutout?cutout=true")
		} else {
			r, err = http.Get(addr + "cutout?cutout=false")
		}
	default:
		return fmt.Errorf("pass either --play, --toogle or --convert or --solo or --cutout")
	}

	if r != nil {
		io.ReadAll(r.Body)
		r.Body.Close()
	}

	return err
}

func (me *App) runConfigDirs(app config.Application) error {
	fmt.Fprintf(os.Stdout, "USER_DIR: %q\nWORKING_DIR: %q\n\n", muskel.USER_DIR, muskel.WORKING_DIR)
	return nil
}

func (me *App) runInspect(app config.Application) error {
	sc, err := muskel.ParseFile(me.InFile.Val, nil)

	if err != nil {
		return err
	}

	fsys, err := sc.AsFS()

	if err != nil {
		return err
	}

	return fsshell.New(fsys, "").Run()
}

func (me *App) runConv(conv *converter.Converter) error {
	file := conv.Dir.Join(conv.File)

	if path.IsDir(file) {
		return fmt.Errorf("is directory: %q", file)
	}

	if path.Ext(file) != muskel.FILE_EXTENSION {
		return fmt.Errorf("no muskel file: %q\n", file)
	}

	cb := me.mkcallback(conv)
	return cb(conv.Dir.String(), conv.Dir.Join(conv.File).String())
}

func (me *App) mkcallback(conv *converter.Converter) (callback func(dir, file string) error) {
	conv.Checksums = map[string]string{}
	conv.Ignore = map[string]bool{}

	callback = func(dir, file string) error {
		return me.prepareConv(conv, dir, file)
	}
	return
}

func (me *App) prepareConv(conv *converter.Converter, dir, file string) error {

	if filepath.Ext(file) != muskel.FILE_EXTENSION {
		return nil
	}

	if strings.Contains(dir, "muskel-fmt") || strings.Contains(file, "muskel-fmt") {
		return nil
	}

	loc, err := path.ParseFileFromSystem(file)

	if err != nil {
		return err
	}

	filep := loc.String()

	if strings.Contains(filep, "muskel-fmt") {
		return nil
	}

	if conv.Ignore[filep] {
		return nil
	}

	if !me.WatchDir.Val {
		if filep != conv.InFile.String() {
			return nil
		}
	}

	if me.Watch.Val {
		newChecksum := fileCheckSum(filep)

		if newChecksum == conv.Checksums[filep] {
			return nil
		}

		conv.Checksums[filep] = newChecksum
	}

	if me.WatchDir.Val {
		if filep != conv.InFile.String() && conv.Config.Fmt {
			conv.Ignore[filep] = true
			conv.FmtFile(conv.InFile, conv.Config.Params, conv.ScoreOptions()...)
			conv.Ignore[filep] = false
		}
	}

	sc, err := conv.ParseMuskel(conv.InFile)

	if err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		oshelpers.Alert("ERROR:", err)
		return err
	}

	switch me.App.SelectedCommand() {
	case me.smf.command:
		return conv.CmdSMF(sc)
	case me.play.command:
		return conv.CmdPlay(SIGNAL_CHANNEL, sc, me.Watch.Val)
	default:
		return nil
	}
}

func (me *App) startServer(conv *converter.Converter) error {
	me.SetConverterConfigFromArgs(conv)
	s := server.New(conv, conv.Config, conv.Player, conv.InFile, conv.File)
	return s.Start(me.server.Address.Val)
}

func (me *App) SetConverterConfigFromArgs(conv *converter.Converter) {
	conv.Config.PrintBarComments = true
	conv.Config.Flow = me.Flow.Val
	if me.Sketch.IsSet() && me.Sketch.Val != "=SCORE" {
		conv.Config.Sketch = me.Sketch.Val
	} else {
		conv.Config.Sketch = ""
	}

	conv.Config.Debug = me.Debug.Val

	if me.Pattern.IsSet() {
		conv.Config.Pattern = me.Pattern.Val
	} else {
		conv.Config.Pattern = ""
	}
	conv.Config.KeepEmptyLines = me.KeepEmptyLines.Val
	if me.Params.IsSet() {
		conv.Config.Params = me.Params.Val
	} else {
		conv.Config.Params = nil
	}

	if me.UnrollFile.IsSet() {
		conv.Config.UnrollFile = me.UnrollFile.Val
	}

	conv.Config.Fmt = me.Fmt.Val

	if muskel.FILE_EXTENSION == ".csv" {
		conv.Config.CSV = me.CSVSeparator.Val
		// never fmt when in csv
		conv.Config.Fmt = false
	}

	if muskel.FILE_EXTENSION == ".xlsx" {
		conv.Config.XLSX = true
		// never fmt when in csv
		conv.Config.Fmt = false
	}

	if me.Solo.IsSet() {
		sg := me.Solo.Val
		if sg > 0 {
			conv.Config.SoloGroup = uint(sg)
		}
	}

	if me.CutOut.Val {
		conv.Config.CutOut = true
	}

	conv.Config.SMF.ExportImage = me.smf.ExportImage.Val
	conv.Config.SMF.ExportScore = me.smf.ExportScore.Val
	conv.Config.SMF.PDFfile = me.smf.PDFfile.Val
	conv.Config.SMF.Watch = me.Watch.Val
	conv.Config.SMF.Open = me.smf.Open.Val

}

func (me *App) runServer(conv *converter.Converter) error {
	if conv.Player == nil {
		return fmt.Errorf("player must not be nil")
	}

	var stopFunc = conv.Player.StopFunc()

	if conv.Player.PlayToPort {
		go conv.Player.PlayThroughPort()
	} else {
		audioOut, err := smfplayer.OpenAudioOut(44100)

		if err != nil {
			return err
		}

		defer audioOut.Close()
		go conv.Player.PlayAsService(audioOut)
	}

	go func() {
		err := me.startServer(conv)
		if err != nil {
			fmt.Printf("ERROR: can't start server: %s\n", err.Error())
			stopFunc()
			os.Exit(1)
		}
	}()

	// ctrl+c
	<-SIGNAL_CHANNEL
	fmt.Println("\n--interrupted!")
	stopFunc()
	os.Exit(0)
	return nil
}

func (me *App) watch(conv *converter.Converter, doWatchDir bool, inFile path.Local, sleepingTimeMS int) error {
	if doWatchDir {
		fmt.Printf("watching %q\n", conv.Dir)
	} else {
		fmt.Printf("watching %q\n", inFile.ToSystem())
	}

	r := runfunc.New(
		path.ToSystem(conv.Dir),
		me.mkcallback(conv),
		runfunc.Sleep(time.Millisecond*time.Duration(sleepingTimeMS)),
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

func (me *App) runTemplate(app config.Application) error {
	if me.template.File.IsSet() && me.template.File.Val != "" {
		return me.printTemplateFile(muskel.USER_DIR.Join(me.template.File.Val + muskel.FILE_EXTENSION))
	} else {
		return me.printTemplateDir(muskel.USER_DIR)
	}
}

func (me *App) printTemplateFile(p path.Local) error {
	if !rootfs.Exists(p) {
		return fmt.Errorf("no such template: %q", p)
	}

	fl, err := rootfs.OpenReader(p)

	if err != nil {
		return fmt.Errorf("can't open template %q: %s", p, err.Error())
	}

	defer fl.Close()
	io.Copy(os.Stdout, fl)
	return nil
}

func (me *App) printTemplateDir(dir path.Local) error {
	fls, err := rootfs.ReadDirPaths(dir)
	if err != nil {
		return fmt.Errorf("can't read user directory %q: %s", dir, err.Error())
	}

	if len(fls) == 0 {
		fmt.Fprintf(os.Stdout, "There are no template files stored inside %s\n\n", dir)
		return nil
	}

	fmt.Fprintf(os.Stdout, "The following template files are stored inside %s:\n\n", dir)

	for _, fl := range fls {
		name := fl.String()
		var tistr string
		ti, err := rootfs.ModTime(fl)
		if err == nil {
			tistr = fmt.Sprintf("(modified: %s)", ti.Format(time.RFC822))
		}
		fmt.Fprintf(os.Stdout, "%-24v %s\n", name, tistr)
	}
	return nil
}

func (me *App) runSMFJSON(app config.Application) error {
	bt, err := rootfs.ReadFile(me.smfJSON.File.Val)

	if err != nil {
		return err
	}

	sm, err := smf.ReadFrom(bytes.NewReader(bt))

	if err != nil {
		return err
	}

	bt, err = sm.MarshalJSONIndent()

	if err != nil {
		return err
	}

	var targetFile path.Local

	if me.smfJSON.OutFile.IsSet() {
		targetFile = me.smfJSON.OutFile.Val
	} else {
		fname := path.Name(me.smfJSON.File.Val)
		fname = strings.Replace(fname, ".mid", "", 1)
		fname = fname + ".json"
		targetFile = path.MustWD().Join(items.MIDI_SAMPLES_DIR.String(), fname)
	}

	return rootfs.WriteFile(targetFile, bt, false)
}

func (me *App) runPorts(app config.Application) error {
	outs := midi.GetOutPorts()
	fmt.Fprintln(os.Stdout, "MIDI outputs")
	for _, out := range outs {
		fmt.Fprintf(os.Stdout, "[%v] %s\n", out.Number(), out.String())
	}
	return nil
}

func (me *App) run(app config.Application) error {
	if !me.Current.Val {
		// check, if muskel_version.txt is inside the current dir
		// If the version does not match, it delegates to the versioned muskel command (e.g. muskel_1_1_33.exe)
		// checkMuskelVersion exits if appropriate
		v := version.BuildVersion()
		oshelpers.CheckMuskelVersion(v.String(), muskel.MUSKEL_VERSION_FILE, me.InFile.Val.Dir())
	}

	// when using markdown files, set the file extension properly (affects includes!)
	switch path.Ext(me.InFile.Val) {
	case ".md":
		muskel.FILE_EXTENSION = ".md"
	case ".csv":
		muskel.FILE_EXTENSION = ".csv"
	case ".xlsx":
		muskel.FILE_EXTENSION = ".xlsx"
	default:
		muskel.FILE_EXTENSION = ".mskl"
	}

	// listen for ctrl+c
	go signal.Notify(SIGNAL_CHANNEL, os.Interrupt)

	player, err := me.newPlayer()
	if err != nil {
		return err
	}

	var audioOut *smfplayer.AudioOut

	if me.smf.Open.Val || (player.Enabled && !player.PlayToPort) {
		fmt.Println("opening audioout")
		audioOut, err = smfplayer.OpenAudioOut(44100)

		if err != nil {
			return err
		}

		defer audioOut.Close()
	}

	conv := converter.New(player, me.InFile.Val, me.Watch.Val, audioOut)
	me.SetConverterConfigFromArgs(conv)

	if me.App.SelectedCommand() == me.server.command {
		// the server keeps running, until ctrl+c
		return me.runServer(conv)
	}

	if me.Watch.Val {
		// the watcher keeps running, until ctrl+c
		return me.watch(conv, me.WatchDir.Val, me.InFile.Val, me.SleepingTime.Val)
	}

	// just one simple conversion / play; clean exit on ctrl+c
	return me.runConv(conv)
}

func (me *App) init() *App {
	me.App = config.New("muskel", me.run).
		WithAsciiArt().
		WithDescription("muskel is a musical sketch language")

	me.InFile.RegExp = regexp.MustCompile(`\.(md|MD|mskl|MSKL)$`)
	me.OutFile.RegExp = regexp.MustCompile(`\.(mid|MID|midi|MIDI)$`)
	me.smf.PDFfile.RegExp = regexp.MustCompile(`\.(pdf|PDF)$`)
	me.play.WaveFile.RegExp = regexp.MustCompile(`\.(wav|WAV)$`)

	me.AddArg("file", &me.InFile).
		WithHelp("path of the muskel file").
		WithRequired().
		WithDefault("./main.mskl")

	me.AddFlag("csvsep", &me.CSVSeparator).
		WithHelp("uses the given string as csv seperator.").
		WithDefault(";")

	me.AddFlag("current", &me.Current).
		WithHelp("use the current version of the muskel command and ignore the " + muskel.MUSKEL_VERSION_FILE + " file").
		WithDefault("false").
		WithShortFlag('c')

	me.AddFlag("sketch", &me.Sketch).
		WithHelp("name of the sketch table").
		WithShortFlag('s').
		WithDefault("=SCORE")

	me.AddFlag("flow", &me.Flow).
		WithHelp("flow mode; sets sketch to ! and pattern to !").
		WithDefault("false")

	me.AddFlag("params", &me.Params).
		WithHelp("parameters passed to the sketch. params must have the syntax [trackname]#[no]:[value] where no is the params number, e.g. [\"voc#2:c#'\",\"piano#1:D\"]").
		WithShortFlag('p').
		WithDefault("[]")

	me.AddFlag("pattern", &me.Pattern).
		WithHelp("pattern to be used exclusively").
		WithShortFlag('t').
		WithDefault("")

	me.AddFlag("fmt", &me.Fmt).
		WithHelp("format the muskel file (overwrites the input file)")

	me.AddFlag("keepemptylines", &me.KeepEmptyLines).
		WithHelp("don't remove empty lines from the score").
		WithShortFlag('k').
		WithDefault("true")

	me.AddFlag("out", &me.OutFile).
		WithHelp("path of the output file (SMF). If it includes the placeholder %s, that will be replaced by the File property of the corresponding track").
		WithShortFlag('o')

	me.AddFlag("trackfiles", &me.TrackFiles).
		WithHelp(`sets out to '%%s' in order to write to the file names as given in the track properties`)

	me.AddFlag("unroll", &me.UnrollFile).
		WithHelp("unroll the source to the given file name").
		WithShortFlag('u')

	me.AddFlag("debug", &me.Debug).
		WithHelp("print debug messages")

	sleepFlag := me.AddFlag("sleep", &me.SleepingTime).
		WithHelp("sleeping time between invocations (in milliseconds)").
		WithDefault("10")

	cutOutFlag := me.AddFlag("cutout", &me.CutOut).
		WithHelp("limit the score to the innermost cutout section (indicated by > and _ after the hash in the score)").
		WithDefault("false")

	soloFlag := me.AddFlag("solo", &me.Solo).
		WithHelp("the tracks of the given solo group should be used exclusively (0 = use all tracks)")

	watchFlag := me.AddFlag("watch", &me.Watch).
		WithHelp("watch for changes of the file and act on each change").
		WithShortFlag('w')

	dirFlag := me.AddFlag("dir", &me.WatchDir).
		WithHelp("watch for changes in the current directory (not just for the input file)").
		WithShortFlag('d')

	{
		me.smf.command = me.App.AddCommand("smf", me.run).
			WithHelp("convert a muskel file to Standard MIDI file format (SMF)")

		me.smf.command.AddArg("file", &me.InFile).
			WithHelp("path of the muskel file").
			WithRequired().
			WithDefault("./main.mskl")

		me.smf.command.AddFlag("ticks", &me.smf.ResolutionTicks).
			WithHelp("resolution of SMF file in ticks").
			WithDefault("960")

		me.smf.command.AddFlag("image", &me.smf.ExportImage).
			WithHelp("export to a PNG image (via smfimage)").
			WithShortFlag('i')

		me.smf.command.AddFlag("score", &me.smf.ExportScore).
			WithHelp("export the score to a PDF (via lilypond)").
			WithShortFlag('l')

		me.smf.command.AddFlag("pdf-file", &me.smf.PDFfile).
			WithHelp("pdf file to export the score to").
			WithDefault("score.pdf")

		me.smf.command.AddFlag("open", &me.smf.Open).
			WithHelp("open the output file with default program")
	}

	{
		me.App.AddCommand("ports", me.runPorts).
			WithHelp("show midi out ports").
			SkipAllBut()
	}

	{
		smfjson := me.App.AddCommand("smfjson", me.runSMFJSON).
			WithHelp("convert a SMF file to a JSON file that can be edited and used as a 'midisample'").
			SkipAllBut()

		smfjson.AddFlag("outfile", &me.smfJSON.OutFile).
			WithHelp("the json file output file")

		smfjson.AddArg("file", &me.smfJSON.File).
			WithHelp("the SMF file").
			WithRequired()
	}

	{
		template := me.App.AddCommand("template", me.runTemplate).
			WithHelp("show global template files").
			SkipAllBut()

		template.AddArg("templatefile", &me.template.File).
			WithHelp("file name of the template file that should be shown. If no file is given, the list of available files is shown.")
	}

	{
		me.play.command = me.App.AddCommand("play", me.run).
			WithHelp("play a muskel file after midi conversion via GM").
			Skip(watchFlag).Skip(dirFlag).Skip(sleepFlag)

		me.play.command.AddArg("file", &me.InFile).
			WithHelp("path of the muskel file").
			WithRequired().
			WithDefault("./main.mskl")

		me.play.command.AddFlag("wav-file", &me.play.WaveFile).
			WithHelp("pass a wav file, if the output should be rendered to a wav-file instead of being played")

		me.play.command.AddFlag("out-port", &me.play.OutPort).
			WithHelp("pass a midi outt port, if the output should be played to an external GM-synthesizer")
	}

	{
		me.server.command = me.App.AddCommand("server", me.run).
			WithHelp("start command server").
			Skip(watchFlag).Skip(dirFlag)

		me.server.command.AddArg("file", &me.InFile).
			WithHelp("path of the muskel file").
			WithRequired().
			WithDefault("./main.mskl")

		me.server.command.AddFlag("addr", &me.server.Address).
			WithHelp("address of the server").
			WithDefault("localhost:8800")

		me.server.command.AddFlag("cmd", &me.server.PlayProgram).
			WithHelp("command to execute when playing (fluidsynth,timidity,audacious,auto or midi-ports (port:4) or custom (pass $_file variable))").
			WithDefault("auto")
	}

	{
		me.App.AddCommand("dirs", me.runConfigDirs).
			WithHelp("show configuration dirs").
			SkipAllBut()
	}

	{
		client := me.App.AddCommand("client", me.run).
			WithHelp("client to the server. allows to send play/stop, toggle and convert commands").
			SkipAllBut(soloFlag, cutOutFlag)

		client.AddFlag("addr", &me.client.ServerAddress).
			WithHelp("address of the server").
			WithDefault("localhost:8800")

		client.AddFlag("toggle", &me.client.Toggle).
			WithHelp("toggle between play/stop playing the current file").
			WithShortFlag('t')

		client.AddFlag("play", &me.client.Play).
			WithHelp("play (true) or stop (false) playing the current file").
			WithShortFlag('p')

		client.AddFlag("convert", &me.client.Convert).
			WithHelp("converts the current file to smf").
			WithShortFlag('c')
	}

	{
		cmdInspect := me.App.AddCommand("inspect", me.runInspect).
			WithHelp("inspect muskel project").
			SkipAllBut()

		cmdInspect.AddArg("file", &me.InFile).
			WithHelp("path of the muskel file").
			WithRequired().
			WithDefault("./main.mskl")
	}

	return me
}
