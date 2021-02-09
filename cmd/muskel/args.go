package main

import (
	"gitlab.com/gomidi/muskel"
	"gitlab.com/metakeule/config"
)

type args struct {
	*config.Config
	File                config.StringGetter
	IgnoreMuskelVersion config.BoolGetter
	Sketch              config.StringGetter
	Flow                config.BoolGetter
	Solo                config.Int32Getter
	CutOut              config.BoolGetter
	Params              config.JSONGetter
	Pattern             config.StringGetter
	Fmt                 config.BoolGetter
	KeepEmptyLines      config.BoolGetter
	Watch               config.BoolGetter
	WatchDir            config.BoolGetter
	OutFile             config.StringGetter
	TrackFiles          config.BoolGetter
	SleepingTime        config.Int32Getter
	UnrollFile          config.StringGetter
	Debug               config.BoolGetter

	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
}

var ARGS = &args{}

func init() {
	ARGS.init()
}

func (a *args) init() {
	ARGS.Config = CONFIG

	// the main argument set
	a.File = a.LastString("file", "path of the muskel file", config.Required)
	a.IgnoreMuskelVersion = a.NewBool("current", "use the current version of the muskel command and ignore the "+muskel.MUSKEL_VERSION_FILE+" file", config.Default(false), config.Shortflag('c'))
	a.Sketch = a.NewString("sketch", "name of the sketch table", config.Shortflag('s'), config.Default("=SCORE"))
	a.Flow = a.NewBool("flow", "flow mode; sets sketch to ! and pattern to !", config.Default(false))
	a.CutOut = a.NewBool("cutout", "limit the score to the innermost cutout section (indicated by > and _ after the hash in the score)", config.Default(false))
	a.Solo = a.NewInt32("solo", "the tracks of the given solo group should be used exclusively (0 = use all tracks)")
	a.Params = a.NewJSON("params", "parameters passed to the sketch. params must have the syntax [trackname]#[no]:[value] where no is the params number, e.g. [\"voc#2:c#'\",\"piano#1:D\"]", config.Shortflag('p'), config.Default("[]"))
	a.Pattern = a.NewString("pattern", "pattern to be used exclusively", config.Shortflag('t'), config.Default(""))
	a.Fmt = a.NewBool("fmt", "format the muskel file (overwrites the input file)")
	a.KeepEmptyLines = a.NewBool("keepemptylines", "don't remove empty lines from the score", config.Shortflag('k'))
	a.Watch = a.NewBool("watch", "watch for changes of the file and act on each change", config.Shortflag('w'))
	a.WatchDir = a.NewBool("dir", "watch for changes in the current directory (not just for the input file)", config.Shortflag('d'))
	a.OutFile = a.NewString("out", "path of the output file (SMF). If it includes the placeholder %s, that will be replaced by the File property of the corresponding track", config.Shortflag('o'))
	a.TrackFiles = a.NewBool("trackfiles", "sets out to '%s' in order to write to the file names as given in the track properties")
	a.SleepingTime = a.NewInt32("sleep", "sleeping time between invocations (in milliseconds)", config.Default(int32(10)))
	a.UnrollFile = a.NewString("unroll", "unroll the source to the given file name", config.Shortflag('u'))
	a.Debug = a.NewBool("debug", "print debug messages")

	//argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
}
