package midisampler

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"text/template"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

var fnMap = template.FuncMap{
	"zero": func(delta int) string {
		return fmt.Sprintf(`
{
	"data": "00",
	"delta": %v,
	"type": "seqdata",
	"description": "zero point"
},`, delta)
	},
}

/*
{
	"data": "00",
	"delta": 24,
	"type": "seqdata"
},
*/

func ZeroMessage() smf.Message {
	bt, err := hex.DecodeString("00")

	if err != nil {
		panic(err.Error())
	}

	return smf.MetaSequencerData(bt)
}

type logger struct {
	bytes.Buffer
}

func (l *logger) Printf(format string, args ...any) {
	fmt.Fprintf(&l.Buffer, "MIDISAMPLE reader: "+format, args...)
}

// RunJSONTemplate replaces the placeholders in the midi template, returns the first track, where the ticks have been
// converted to the target ticksPerQuarterNote the negative offset has been calculated by the position of the zero.
// the track stops at the position of lenTicks after zero any playing notes will get a noteoff there
func RunJSONTemplate(jsonStr string, ticksPerQuarterNote uint16, trackno int, lenTicks uint64, placeholderMap map[string]any) (track smf.Track, negOffset uint64, err error) {
	var t *template.Template
	t, err = template.New("json-template").Funcs(fnMap).Parse(jsonStr)

	if err != nil {
		err = fmt.Errorf("template parsing error: %s\n", err.Error())
		return
	}

	var wr bytes.Buffer

	err = t.Execute(&wr, placeholderMap)

	if err != nil {
		err = fmt.Errorf("template executing error: %s\n", err.Error())
		return
	}

	var sm smf.SMF
	err = sm.UnmarshalJSON(wr.Bytes())

	if err != nil {
		err = fmt.Errorf("json unmarshaling error: %s\n", err.Error())
		fmt.Println(wr.String())
		return
	}

	if len(sm.Tracks) <= trackno {
		err = fmt.Errorf("track %v not found", trackno)
		return
	}

	mt, hasMT := sm.TimeFormat.(smf.MetricTicks)

	if !hasMT {
		err = fmt.Errorf("midi 'sample' has no metric ticks time format")
		return
	}

	resTicksPerQuarternote := mt.Resolution()

	convertTicks := func(tpq uint32) uint32 {
		return (tpq * uint32(ticksPerQuarterNote) / uint32(resTicksPerQuarternote))
	}

	srcTrack := sm.Tracks[trackno]

	var deltaAdd uint32
	var foundZero bool
	var calcLength uint64
	var keyTracker = [16][128]bool{}

	defer track.Close(0)

	var channel uint8 = 100 // means: not yet found, we want all channel messages to be of the same channel, otherwise it is an error

	for _, ev := range srcTrack {
		delta := convertTicks(ev.Delta) + deltaAdd
		deltaAdd = 0

		var bt []byte
		if ev.Message.GetMetaSeqData(&bt) && hex.EncodeToString(bt) == "00" {
			if foundZero {
				err = fmt.Errorf("has more than one ZERO point")
				return
			}
			negOffset += uint64(delta)
			foundZero = true
			deltaAdd = convertTicks(ev.Delta)
			continue
		}

		switch {
		case !foundZero:
			negOffset += uint64(delta)
		case foundZero:
			if (calcLength + uint64(delta)) > lenTicks {
				// stop all notes
				track.Add(uint32(lenTicks-calcLength), smf.MetaText("autostop"))

				for ch, keys := range keyTracker {
					for key, isOn := range keys {
						if isOn {
							track.Add(0, midi.NoteOff(uint8(ch), uint8(key)))
						}
					}
				}

				return
			}
			// lenTicks
			calcLength += uint64(delta)
		}

		var ch, key, vel uint8

		if ev.Message.GetChannel(&ch) {
			if channel == 100 { // not yet set
				channel = ch
			}

			if channel != ch {
				err = fmt.Errorf("can't have channel messages of more than one channel in midisample file: found channel %v and %v", channel+1, ch+1)
				return
			}
		}

		switch {

		case ev.Message.GetNoteStart(&ch, &key, &vel):
			//	fmt.Printf("setting keyTracker for ch: %v key: %v\n", ch, key)
			keyTracker[ch][key] = true

		case ev.Message.GetNoteEnd(&ch, &key):
			//	fmt.Printf("unsetting keyTracker for ch: %v key: %v\n", ch, key)
			keyTracker[ch][key] = false

		}

		track.Add(delta, ev.Message)
	}

	return
}
