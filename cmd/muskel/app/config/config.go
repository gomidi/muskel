package config

import "gitlab.com/golang-utils/fs/path"

type Config struct {
	PrintBarComments bool
	Flow             bool
	CutOut           bool
	Fmt              bool
	Debug            bool
	KeepEmptyLines   bool
	XLSX             bool
	SoloGroup        uint
	Sketch           string
	Pattern          string
	UnrollFile       string
	CSV              string
	Params           []string

	SMF struct {
		ExportImage bool
		ExportScore bool
		Watch       bool
		Open        bool
		PDFfile     path.Local
	}
}
