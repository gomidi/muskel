package items

import (
	"gitlab.com/gomidi/midi"
)

type SMFWriter interface {
	RestoreTimeline()
	Position() uint64
	Delta() uint32
	Aftertouch(uint8) error
	BackupTimeline()
	Plan(uint32, uint32, uint32, midi.Message)
	Channel() uint8
	ControlChange(uint8, uint8) error
	Pitchbend(int16) error
	PolyAftertouch(uint8, uint8) error
	Forward(uint32, uint32, uint32)
	NoteOn(uint8, uint8) error
	NoteOff(uint8) error
	Lyric(string) error
	Undefined(typ byte, bt []byte) error
	Ticks256th() uint32
	Ticks16th() uint32
	FinishPlanned()
	SetChannel(uint8)
	Program(string) error
	TrackSequenceName(string) error
	Instrument(string) error
	ProgramChange(uint8) error
	PitchBendSensitivityRPN(uint8, uint8) error
	Copyright(string) error
	Cuepoint(string) error
	Marker(string) error
	EndOfTrack() error
	TempoBPM(float64) error
	Meter(uint8, uint8) error
}
