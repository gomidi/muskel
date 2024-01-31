package main

import "gitlab.com/golang-utils/config/v2"

type smfCmd struct {
	*config.Config
	ResolutionTicks config.IntGetter
	ExportImage     config.BoolGetter
	ExportScore     config.BoolGetter
}

var SMF = smfCmd{}

func init() {
	SMF.init()
}

func (s *smfCmd) init() {
	s.Config = CONFIG.Command("smf", "convert a muskel file to Standard MIDI file format (SMF)")
	s.ResolutionTicks = s.Int("ticks", "resolution of SMF file in ticks", config.Default(int32(960)))
	s.ExportImage = s.Bool("image", "export to a PNG image (via smfimage)", config.Shortflag('i'))
	s.ExportScore = s.Bool("score", "export the score to a PDF (via lilypond)", config.Shortflag('l'))
}
