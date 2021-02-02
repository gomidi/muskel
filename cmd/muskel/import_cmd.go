package main

import (
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/metakeule/config"
)

type importCmd struct {
	*config.Config
	Source      config.StringGetter
	OutFile     config.StringGetter
	MonoTracks  config.StringGetter
	OnsetTracks config.StringGetter
}

var IMPORT = &importCmd{}

func init() {
	IMPORT.init()
}

func (s *importCmd) init() {
	s.Config = CONFIG.MustCommand("import", "convert a Standard MIDI file to a muskel file").SkipAllBut()
	s.Source = s.LastString("midifile", "the source SMF/MIDI file.", config.Required)
	s.OutFile = s.NewString("out", "the target muskel file.", config.Shortflag('o'), config.Required)
	s.MonoTracks = s.NewString("mono", "tracks that are considered to be monophon (e.g. 0,4,5)", config.Shortflag('m'))
	s.OnsetTracks = s.NewString("onset", "tracks that only have onset data (e.g. drum notes of or keyswitches) (e.g. 0,4,5)", config.Shortflag('r'))
}

func (i *importCmd) run() error {
	var monoTracks []int
	var onsetTracks []int

	if i.MonoTracks.IsSet() {
		ms := strings.Split(i.MonoTracks.Get(), ",")

		for _, m := range ms {
			i, errI := strconv.Atoi(strings.TrimSpace(m))
			if errI == nil {
				monoTracks = append(monoTracks, i)
			}
		}
	}

	if i.OnsetTracks.IsSet() {
		ms := strings.Split(i.OnsetTracks.Get(), ",")

		for _, m := range ms {
			i, errI := strconv.Atoi(strings.TrimSpace(m))
			if errI == nil {
				onsetTracks = append(onsetTracks, i)
			}
		}
	}

	return muskel.Import(i.Source.Get(), i.OutFile.Get(), monoTracks, onsetTracks)
}
