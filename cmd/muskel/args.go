package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"gitlab.com/golang-utils/config/v3"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"
	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/items"
)

var (
	ARGS           = &args{}
	APP            = config.New("muskel", ARGS.run).WithAsciiArt().WithDescription("muskel is a musical sketch language")
	SIGNAL_CHANNEL = make(chan os.Signal, 10)
)

type args struct {
	*config.App
	InFile              config.FileLocal
	CSVSeparator        config.String
	IgnoreMuskelVersion config.Bool
	Sketch              config.String
	Flow                config.Bool
	Solo                config.Int
	CutOut              config.Bool
	Params              config.JSON[[]string]
	Pattern             config.String
	Fmt                 config.Bool
	KeepEmptyLines      config.Bool
	Watch               config.Bool
	WatchDir            config.Bool
	OutFile             config.FileLocal
	TrackFiles          config.Bool
	SleepingTime        config.Int
	UnrollFile          config.String
	Debug               config.Bool

	pdf struct {
		PDFfile      config.FileLocal
		Experimental config.Bool
	}

	smf struct {
		command         *config.Command
		ResolutionTicks config.Int
		ExportImage     config.Bool
		ExportScore     config.Bool
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
		Program config.Select[string] // config.String // fluidsynth,timidity,audacious,auto
	}

	server struct {
		command      *config.Command
		Address      config.String
		PlayProgram  config.String
		playershared *Player
	}

	client struct {
		ServerAddress config.String
		Toggle        config.Bool
		Play          config.Bool
		Convert       config.Bool
	}

	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
}

func (a *args) runClient() error {
	var r *http.Response
	var err error
	addr := "http://" + a.client.ServerAddress.Val + "/"
	switch {
	case a.client.Play.IsSet():
		if a.client.Play.Val {
			r, err = http.Get(addr + "play")
		} else {
			r, err = http.Get(addr + "stop")
		}
	case a.client.Convert.IsSet():
		if a.client.Convert.Val {
			r, err = http.Get(addr + "convert")
		}
	case a.client.Toggle.IsSet():
		if a.client.Toggle.Val {
			r, err = http.Get(addr + "toggle")
		}
	case a.Solo.IsSet():
		r, err = http.Get(addr + fmt.Sprintf("solo?solo=%v", a.Solo.Val))
	case a.CutOut.IsSet():
		if a.CutOut.Val {
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

func (a *args) runConfigDirs(app config.Application) error {
	fmt.Fprintf(os.Stdout, "USER_DIR: %q\nWORKING_DIR: %q\n\n", muskel.USER_DIR, muskel.WORKING_DIR)
	return nil
}

func (a *args) runServer(cbr *converter) error {
	if cbr.player == nil {
		return fmt.Errorf("player must not be nil")
	}

	var stopFunc = func() {
		cbr.player.stopPlayer <- true
		<-cbr.player.playerStopped
	}

	if cbr.player.playToPort {
		go cbr.player.playThroughPort()
	} else {
		go cbr.player.playWithProgram()
	}

	go func() {
		err := startServer(cbr)
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

func (a *args) runTemplate(app config.Application) error {
	if a.template.File.IsSet() && a.template.File.Val != "" {
		return a.printTemplateFile(muskel.USER_DIR.Join(a.template.File.Val + muskel.FILE_EXTENSION))
	} else {
		return a.printTemplateDir(muskel.USER_DIR)
	}
}

func (a *args) printTemplateFile(p path.Local) error {
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

func (a *args) printTemplateDir(dir path.Local) error {
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

func (a *args) runSMFJSON(app config.Application) error {
	bt, err := rootfs.ReadFile(a.smfJSON.File.Val)

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

	if a.smfJSON.OutFile.IsSet() {
		targetFile = a.smfJSON.OutFile.Val
	} else {
		fname := path.Name(a.smfJSON.File.Val)
		fname = strings.Replace(fname, ".mid", "", 1)
		fname = fname + ".json"
		targetFile = path.MustWD().Join(items.MIDI_SAMPLES_DIR.String(), fname)
	}

	return rootfs.WriteFile(targetFile, bt, false)
}

func (a *args) runPorts(app config.Application) error {
	outs := midi.GetOutPorts()
	fmt.Fprintln(os.Stdout, "MIDI outputs")
	for _, out := range outs {
		fmt.Fprintf(os.Stdout, "[%v] %s\n", out.Number(), out.String())
	}
	return nil
}

func init() {
	ARGS.init()
}

func (a *args) run(app config.Application) error {
	if !a.IgnoreMuskelVersion.Val {
		// check, if muskel_version.txt is inside the current dir
		// If the version does not match, it delegates to the versioned muskel command (e.g. muskel_1_1_33.exe)
		// checkMuskelVersion exits if appropriate
		v := version.BuildVersion()
		checkMuskelVersion(v.String(), muskel.MUSKEL_VERSION_FILE, a)
	}

	// when using markdown files, set the file extension properly (affects includes!)
	switch path.Ext(a.InFile.Val) {
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

	player, err := newPlayer()
	if err != nil {
		return err
	}

	conv := newConverter(player, a)

	if a.App.SelectedCommand() == a.server.command {
		// the server keeps running, until ctrl+c
		return a.runServer(conv)
	}

	if a.Watch.Val {
		// the watcher keeps running, until ctrl+c
		return watch(conv, a)
	}

	// just one simple conversion / play; clean exit on ctrl+c
	return conv.run()
}

func (a *args) runPDF(app config.Application) error {
	if a.pdf.Experimental.Val {
		return muskel.PDFScore(a.InFile.Val, nil, a.pdf.PDFfile.Val)
	}

	dir, err := rootfs.MkDirTmp("muskel-pdf*")

	if err != nil {
		return err
	}

	file := dir.Join("muskel-lilypond.mid")

	defer rootfs.Delete(dir, true)

	err = muskel.Convert(a.InFile.Val, nil, file)

	if err != nil {
		return err
	}

	return lilypond.MIDI2PDF(file.ToSystem(), a.pdf.PDFfile.Val.ToSystem(), true)
}

func (a *args) init() {
	a.App = APP

	a.InFile.RegExp = regexp.MustCompile(`\.(md|MD|mskl|MSKL)$`)
	a.OutFile.RegExp = regexp.MustCompile(`\.(mid|MID|midi|MIDI)$`)
	a.pdf.PDFfile.RegExp = regexp.MustCompile(`\.(pdf|PDF)$`)

	// the main argument set
	a.AddArg("file", &a.InFile).WithHelp("path of the muskel file").WithRequired()
	a.AddFlag("csvsep", &a.CSVSeparator).WithHelp("uses the given string as csv seperator.").WithDefault(";")
	a.AddFlag("current", &a.IgnoreMuskelVersion).WithHelp("use the current version of the muskel command and ignore the " + muskel.MUSKEL_VERSION_FILE + " file").WithDefault("false").WithShortFlag('c')
	a.AddFlag("sketch", &a.Sketch).WithHelp("name of the sketch table").WithShortFlag('s').WithDefault("=SCORE")
	a.AddFlag("flow", &a.Flow).WithHelp("flow mode; sets sketch to ! and pattern to !").WithDefault("false")
	cutOutFlag := a.AddFlag("cutout", &a.CutOut).WithHelp("limit the score to the innermost cutout section (indicated by > and _ after the hash in the score)").WithDefault("false")
	soloFlag := a.AddFlag("solo", &a.Solo).WithHelp("the tracks of the given solo group should be used exclusively (0 = use all tracks)")
	a.AddFlag("params", &a.Params).WithHelp("parameters passed to the sketch. params must have the syntax [trackname]#[no]:[value] where no is the params number, e.g. [\"voc#2:c#'\",\"piano#1:D\"]").WithShortFlag('p').WithDefault("[]")
	a.AddFlag("pattern", &a.Pattern).WithHelp("pattern to be used exclusively").WithShortFlag('t').WithDefault("")
	a.AddFlag("fmt", &a.Fmt).WithHelp("format the muskel file (overwrites the input file)")
	a.AddFlag("keepemptylines", &a.KeepEmptyLines).WithHelp("don't remove empty lines from the score").WithShortFlag('k').WithDefault("true")
	watchFlag := a.AddFlag("watch", &a.Watch).WithHelp("watch for changes of the file and act on each change").WithShortFlag('w')
	dirFlag := a.AddFlag("dir", &a.WatchDir).WithHelp("watch for changes in the current directory (not just for the input file)").WithShortFlag('d')
	a.AddFlag("out", &a.OutFile).WithHelp("path of the output file (SMF). If it includes the placeholder %s, that will be replaced by the File property of the corresponding track").WithShortFlag('o')
	a.AddFlag("trackfiles", &a.TrackFiles).WithHelp(`sets out to '%%s' in order to write to the file names as given in the track properties`)
	sleepFlag := a.AddFlag("sleep", &a.SleepingTime).WithHelp("sleeping time between invocations (in milliseconds)").WithDefault("10")
	a.AddFlag("unroll", &a.UnrollFile).WithHelp("unroll the source to the given file name").WithShortFlag('u')
	a.AddFlag("debug", &a.Debug).WithHelp("print debug messages")

	pdf := a.App.AddCommand("pdf", a.runPDF).WithHelp("convert a muskel file to pdf via lilypond").SkipAllBut()
	pdf.AddArg("file", &a.InFile).WithHelp("path of the muskel file").WithRequired()
	pdf.AddFlag("pdf", &a.pdf.PDFfile).WithHelp("pdf file to export to").WithRequired()
	pdf.AddFlag("experiment", &a.pdf.Experimental).WithHelp("use the experimental work in progress lilypond support").WithShortFlag('x')

	a.smf.command = a.App.AddCommand("smf", a.run).WithHelp("convert a muskel file to Standard MIDI file format (SMF)")
	a.smf.command.AddArg("file", &a.InFile).WithHelp("path of the muskel file").WithRequired()
	a.smf.command.AddFlag("ticks", &a.smf.ResolutionTicks).WithHelp("resolution of SMF file in ticks").WithDefault("960")
	a.smf.command.AddFlag("image", &a.smf.ExportImage).WithHelp("export to a PNG image (via smfimage)").WithShortFlag('i')
	a.smf.command.AddFlag("score", &a.smf.ExportScore).WithHelp("export the score to a PDF (via lilypond)").WithShortFlag('l')

	a.App.AddCommand("ports", a.runPorts).WithHelp("show midi out ports").SkipAllBut()

	smfjson := a.App.AddCommand("smfjson", a.runSMFJSON).WithHelp("convert a SMF file to a JSON file that can be edited and used as a 'midisample'").SkipAllBut()
	smfjson.AddFlag("outfile", &a.smfJSON.OutFile).WithHelp("the json file output file")
	smfjson.AddArg("file", &a.smfJSON.File).WithHelp("the SMF file").WithRequired()

	template := a.App.AddCommand("template", a.runTemplate).WithHelp("show global template files").SkipAllBut()
	template.AddArg("templatefile", &a.template.File).WithHelp("file name of the template file that should be shown. If no file is given, the list of available files is shown.")

	a.play.command = a.App.AddCommand("play", a.run).
		WithHelp("play a muskel file via smfplayer (timidity,fluidsynth or audacity) or directly to midi out ports").
		Skip(watchFlag).Skip(dirFlag).Skip(sleepFlag)

	a.play.command.AddArg("file", &a.InFile).WithHelp("path of the muskel file").WithRequired()
	a.play.Program = config.NewSelectWithNew[string]("fluidsynth", "timidity", "audacious", "auto")
	a.play.command.AddFlag("cmd", &a.play.Program).
		WithHelp("command to execute when playing (fluidsynth,timidity,audacious,auto or midi-ports (port:4) or custom (pass $_file variable))").
		WithDefault("auto")

	a.server.command = a.App.AddCommand("server", a.run).WithHelp("start command server").Skip(watchFlag).Skip(dirFlag)
	a.server.command.AddArg("file", &a.InFile).WithHelp("path of the muskel file").WithRequired()
	a.server.command.AddFlag("addr", &a.server.Address).WithHelp("address of the server").WithDefault("localhost:8800")
	a.server.command.AddFlag("cmd", &a.server.PlayProgram).
		WithHelp("command to execute when playing (fluidsynth,timidity,audacious,auto or midi-ports (port:4) or custom (pass $_file variable))").
		WithDefault("auto")

	a.App.AddCommand("dirs", a.runConfigDirs).WithHelp("show configuration dirs").SkipAllBut()

	client := a.App.AddCommand("client", a.run).
		WithHelp("client to the server. allows to send play/stop, toggle and convert commands").
		SkipAllBut(soloFlag, cutOutFlag)

	client.AddFlag("addr", &a.client.ServerAddress).WithHelp("address of the server").WithDefault("localhost:8800")
	client.AddFlag("toggle", &a.client.Toggle).WithHelp("toggle between play/stop playing the current file").WithShortFlag('t')
	client.AddFlag("play", &a.client.Play).WithHelp("play (true) or stop (false) playing the current file").WithShortFlag('p')
	client.AddFlag("convert", &a.client.Convert).WithHelp("converts the current file to smf").WithShortFlag('c')

	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
}
