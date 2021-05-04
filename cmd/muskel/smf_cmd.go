package main

import "gitlab.com/metakeule/config"

type smfCmd struct {
	*config.Config
	ResolutionTicks config.Int32Getter
	ExportImage     config.BoolGetter
	ExportScore     config.BoolGetter
}

var SMF = smfCmd{}

func init() {
	SMF.init()
}

func (s *smfCmd) init() {
	s.Config = CONFIG.MustCommand("smf", "convert a muskel file to Standard MIDI file format (SMF)")
	s.ResolutionTicks = s.NewInt32("ticks", "resolution of SMF file in ticks", config.Default(int32(960)))
	s.ExportImage = s.NewBool("image", "export to a PNG image (via smfimage)", config.Shortflag('i'))
	s.ExportScore = s.NewBool("score", "export the score to a PDF (via lilypond)", config.Shortflag('l'))
}
