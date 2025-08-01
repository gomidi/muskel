package smf

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/rpn_nrpn"
	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/track"
)

type Track struct {
	midiTrack MidiTrack
	smf       *SMF
}

func (iw *Track) trackIntroEvents(tr *track.Track) (evts Events) {
	ch := uint8(tr.MIDIChannel)

	evts = append(evts, &Event{Message: smf.MetaProgram(tr.Name)})
	evts = append(evts, &Event{Message: smf.MetaTrackSequenceName(tr.Name)})
	evts = append(evts, &Event{Message: smf.MetaInstrument(tr.Name)})

	if tr.MIDIBank >= 0 {
		//fmt.Printf("Bank: %v\n", iw.instr.MIDIBank)
		evts = append(evts, &Event{Message: midi.ControlChange(ch, midi.BankSelectMSB, uint8(tr.MIDIBank)).Bytes()})
	}

	if tr.MIDIProgram >= 0 {
		//fmt.Printf("ProgramChange: %v\n", iw.instr.MIDIProgram)
		evts = append(evts, &Event{Message: midi.ProgramChange(ch, uint8(tr.MIDIProgram)).Bytes()})
	}

	if tr.MIDIVolume >= 0 {
		//fmt.Printf("Volume: %v\n", iw.instr.MIDIVolume)
		evts = append(evts, &Event{Message: midi.ControlChange(ch, midi.VolumeMSB, uint8(tr.MIDIVolume)).Bytes()})
	}

	pitchBendRange := uint8(2)
	if tr.PitchbendRange > 0 {
		pitchBendRange = tr.PitchbendRange

	}

	msgs := rpn_nrpn.PitchBendSensitivity(uint8(tr.MIDIChannel), pitchBendRange, 0)

	for _, msg := range msgs {
		evts = append(evts, &Event{Message: msg.Bytes()})
	}

	msgs = rpn_nrpn.RPNReset(uint8(tr.MIDIChannel))

	for _, msg := range msgs {
		evts = append(evts, &Event{Message: msg.Bytes()})
	}

	return
}

func (tw *Track) barOfEvent(ev *items.Event) (bar *items.Bar) {
	if len(tw.smf.score.Bars) == 0 {
		panic("no bars")
	}

	idx := -1

	for _, b := range tw.smf.score.Bars {
		//fmt.Printf("b.Position: %v ev.Position: %v\n", b.Position, ev.Position)
		if b.Position > ev.Position {
			break
		}
		idx++
	}

	if idx < 0 {
		idx = 0
	}

	return tw.smf.score.Bars[idx]
}

func (t *Track) convertSketchEvent(ch uint8, sketchEvent *items.Event, trackDelay int) (evts Events, err error) {
	//fmt.Printf("convertSketchEvent trackdelay: %v\n", trackDelay)
	if sketchEvent == nil || sketchEvent.Item == nil || sketchEvent.Item == items.Hold {
		return nil, nil
	}

	pos := TicksAbsPos(sketchEvent.AbsPosTicks(t.smf.writer.ticks, trackDelay))

	//fmt.Printf("pos: %v, event: %s\n", pos, sketchEvent)

	if sketchEvent.Item == items.Rest {
		evts = append(evts, &Event{Position: pos, StopNotes: true, Message: nil})
		return
	}

	switch v := sketchEvent.Item.(type) {

	case *items.Note:
		vscale := t.midiTrack.VelocityScale
		vel := items.DynamicToVelocity(v.Dynamic, int8(vscale[0]), int8(vscale[1]), int8(vscale[2]), int8(vscale[3]), int8(vscale[4]))

		if vel < 0 {
			vel = t.midiTrack.PrevVel
		}

		t.midiTrack.PrevVel = vel

		vl := uint8(vel + int8(rand.Intn(int(vscale[2]))))

		var n uint8
		n = v.ToMIDI()

		// only add MIDITranspose to Notes not to MIDINotes
		n = uint8(int8(n) + t.midiTrack.Track.MIDITranspose)
		t.midiTrack.PrevKey = n

		if t.midiTrack.noteGlide.active {
			/*
				1. get the distance to the previous note
				2. calc the diff in semitones
				3. make a line of 32ths in steps of the factor
			*/
			distance := int64(math.Round(float64(uint64(pos)-t.midiTrack.noteGlide.startPosition) / float64(t.midiTrack.GlideResolution(t.smf.writer.ticks))))
			noteDiff := int64(n) - int64(t.midiTrack.noteGlide.startNote)

			var pb int16

			t.midiTrack.noteGlide.glideFunc(distance, noteDiff, func(step uint, vl float64) {
				pb = items.HalfTonesToPitchbend(vl, t.midiTrack.Track.PitchbendRange)

				evts = append(evts, &Event{
					Position: TicksAbsPos(uint(t.midiTrack.noteGlide.startPosition + uint64(t.smf.writer.ticks/16)*uint64(step))),
					Message:  midi.Pitchbend(ch, pb).Bytes(),
				})
			})

			t.midiTrack.PrevPitchbend = pb
		}

		var ev *Event
		ev = &Event{Position: pos}

		if !t.midiTrack.noteGlide.active {
			if t.midiTrack.noteGlide.startNote != 0 {
				t.midiTrack.noteGlide.startNote = 0
				evts = append(evts, &Event{Position: pos, Message: midi.Pitchbend(ch, 0).Bytes()})
				t.midiTrack.PrevPitchbend = 0
			}

			if !v.NoteOff {
				ev.Message = midi.NoteOn(ch, n, vl).Bytes()
			}

			if !v.NoteOn && !v.NoteOff {
				ev.Monitor = true
				ev.StopNotes = true
				if v.Dotted != "" {
					l := items.DottedLengths[v.Dotted][1]
					evts = append(evts, &Event{Position: TicksAbsPos(uint(uint16(pos) + (t.smf.writer.ticks * 4 / l))), Message: midi.NoteOff(ch, n).Bytes()})
				}

				if t.midiTrack.noteGlide.noteOffBefore > 0 {
					evts = append(evts, &Event{
						Position: pos,
						Message:  midi.NoteOff(ch, t.midiTrack.noteGlide.noteOffBefore).Bytes()})
					t.midiTrack.noteGlide.noteOffBefore = 0
				}
			}
			evts = append(evts, ev)

		}

		if v.NoteOff {
			ev.Message = midi.NoteOff(ch, n).Bytes()
			evts = append(evts, ev)
		}

		if t.midiTrack.noteGlide.active && !v.NoteOff && !v.NoteOn {
			//  set pitchbend back to 0
			if v.Dotted != "" {
				l := items.DottedLengths[v.Dotted][1]
				evts = append(evts, &Event{
					Position: pos,
					Message:  midi.Pitchbend(ch, t.midiTrack.PrevPitchbend).Bytes()})
				evts = append(evts, &Event{Position: TicksAbsPos(uint16(pos) + (t.smf.writer.ticks * 4 / l)),
					Message: midi.NoteOff(ch, t.midiTrack.noteGlide.startNote).Bytes()})
			}
		}

		if t.midiTrack.noteGlide.active {
			t.midiTrack.noteGlide.active = false
			if v.Dotted == "" {
				t.midiTrack.noteGlide.noteOffBefore = t.midiTrack.noteGlide.startNote
			}
		}

		if v.GlissandoStart && !v.NoteOff && !v.NoteOn {
			t.midiTrack.noteGlide.startPosition = uint64(pos)
			t.midiTrack.noteGlide.startNote = n
			t.midiTrack.noteGlide.active = true
			t.midiTrack.noteGlide.glideFunc = items.LinearGlide
			if v.GlissandoExp {
				t.midiTrack.noteGlide.glideFunc = items.ExponentialGlide
			}
		}

	case *items.MIDINote:
		var ev *Event
		ev = &Event{Position: pos}

		t.midiTrack.noteGlide.active = false
		if !v.NoteOff && !v.NoteOn {
			ev.StopNotes = true
			ev.Monitor = true
			if t.midiTrack.noteGlide.noteOffBefore > 0 {
				evts = append(evts, &Event{
					Position: pos,
					Message:  midi.NoteOff(ch, t.midiTrack.noteGlide.noteOffBefore).Bytes()})
				t.midiTrack.noteGlide.noteOffBefore = 0
			}
		}

		vscale := t.midiTrack.VelocityScale
		vel := items.DynamicToVelocity(v.Dynamic, int8(vscale[0]), int8(vscale[1]), int8(vscale[2]), int8(vscale[3]), int8(vscale[4]))

		if vel < 0 {
			vel = t.midiTrack.PrevVel
		}

		t.midiTrack.PrevVel = vel

		vl := uint8(vel + int8(rand.Intn(int(vscale[2]))))
		n := uint8(v.Note)
		if t.midiTrack.noteGlide.startNote != 0 {
			t.midiTrack.noteGlide.startNote = 0
			evts = append(evts, &Event{
				Position: pos,
				Message:  midi.Pitchbend(ch, 0).Bytes()})
		}

		if !v.NoteOff {
			ev.Message = midi.NoteOn(ch, n, vl).Bytes()
		}

		if !v.NoteOff && !v.NoteOn {
			if v.Dotted != "" {
				l := items.DottedLengths[v.Dotted][1]
				evts = append(evts, &Event{
					Position: TicksAbsPos(uint16(pos) + (t.smf.writer.ticks * 4 / l)),
					Message:  midi.NoteOff(ch, n).Bytes()})
			}
		}

		if v.NoteOff {
			ev.Message = midi.NoteOff(ch, n).Bytes()
		}

		evts = append(evts, ev)

	case *items.Lyric:
		evts = append(evts, &Event{Position: pos, Message: smf.MetaLyric(v.Text)})
	case *items.MIDIProgramChange:
		evts = append(evts, &Event{Position: pos, Message: midi.ProgramChange(ch, v.Value).Bytes()})
	case *items.MIDIAftertouch:
		if t.midiTrack.aftertouchGlide.active {
			distance := int64(math.Round(float64(uint64(pos)-t.midiTrack.aftertouchGlide.startPosition) / float64(t.midiTrack.GlideResolution(t.smf.writer.ticks))))
			diff := int64(v.Value) - int64(t.midiTrack.aftertouchGlide.startValue)

			t.midiTrack.aftertouchGlide.glideFunc(distance, diff, func(step uint, vl float64) {
				vll := math.Round(vl + float64(t.midiTrack.aftertouchGlide.startValue))
				if vll > 127 {
					vll = 127
				}

				if vll < 0 {
					vll = 0
				}

				evts = append(evts,
					&Event{
						Position: TicksAbsPos(t.midiTrack.aftertouchGlide.startPosition) + TicksAbsPos(t.smf.writer.ticks)*TicksAbsPos(step)/16,
						Message:  midi.AfterTouch(ch, uint8(vll)).Bytes(),
					})
			})
			t.midiTrack.aftertouchGlide.active = false
		}

		evts = append(evts,
			&Event{
				Position: pos,
				Message:  midi.AfterTouch(ch, v.Value).Bytes(),
			})

		if len(v.Tilde) > 0 {
			t.midiTrack.aftertouchGlide.startPosition = uint64(pos)
			t.midiTrack.aftertouchGlide.startValue = v.Value
			t.midiTrack.aftertouchGlide.active = true
			t.midiTrack.aftertouchGlide.glideFunc = items.LinearGlide
			if v.Tilde == "~~" {
				t.midiTrack.aftertouchGlide.glideFunc = items.ExponentialGlide
			}
		}

		if v.Dotted != "" && v.Value > 0 {
			l := items.DottedLengths[v.Dotted][1]
			evts = append(evts, &Event{
				Position: TicksAbsPos(uint16(pos) + (t.smf.writer.ticks * 4 / l)),
				Message:  midi.AfterTouch(ch, 0).Bytes()})
		}
	case *items.MIDICC:
		if t.midiTrack.ccGlide.active && v.Controller == t.midiTrack.ccGlide.controller {
			distance := int64(math.Round(float64(uint64(pos)-t.midiTrack.ccGlide.startPosition) / float64(t.midiTrack.GlideResolution(t.smf.writer.ticks))))
			diff := int64(v.Value) - int64(t.midiTrack.ccGlide.startValue)

			t.midiTrack.ccGlide.glideFunc(distance, diff, func(step uint, vl float64) {
				vll := math.Round(vl + float64(t.midiTrack.ccGlide.startValue))
				if vll > 127 {
					vll = 127
				}

				if vll < 0 {
					vll = 0
				}

				evts = append(evts,
					&Event{
						Position: TicksAbsPos(t.midiTrack.ccGlide.startPosition) + TicksAbsPos(t.smf.writer.ticks)*TicksAbsPos(step)/16,
						Message:  midi.ControlChange(ch, t.midiTrack.ccGlide.controller, uint8(vll)).Bytes(),
					})

			})
			t.midiTrack.ccGlide.active = false
		}

		evts = append(evts,
			&Event{
				Position: pos,
				Message:  midi.ControlChange(ch, v.Controller, v.Value).Bytes(),
			})

		if len(v.Tilde) > 0 {
			t.midiTrack.ccGlide.controller = v.Controller
			t.midiTrack.ccGlide.startPosition = uint64(pos)
			t.midiTrack.ccGlide.startValue = v.Value
			t.midiTrack.ccGlide.active = true
			t.midiTrack.ccGlide.glideFunc = items.LinearGlide
			if v.Tilde == "~~" {
				t.midiTrack.ccGlide.glideFunc = items.ExponentialGlide
			}
		}

		if v.Dotted != "" && v.Value > 0 {
			l := items.DottedLengths[v.Dotted][1]
			evts = append(evts, &Event{
				Position: TicksAbsPos(uint16(pos) + (t.smf.writer.ticks * 4 / l)),
				Message:  midi.ControlChange(ch, v.Controller, 0).Bytes()})
		}

	case *items.MIDIPitchbend:
		if t.midiTrack.pitchbendGlide.active {
			distance := int64(math.Round(float64(uint64(pos)-t.midiTrack.pitchbendGlide.startPosition) / float64(t.midiTrack.GlideResolution(t.smf.writer.ticks))))
			diff := int64(v.Value+8191) - int64(t.midiTrack.pitchbendGlide.startValue+8191)
			// fmt.Printf("%v - %v = pb diff: %v\n", int8(v.value+8191), int8(iw.startPBGlissandoValue+8191), diff)

			t.midiTrack.pitchbendGlide.glideFunc(distance, diff, func(step uint, vl float64) {
				vll := math.Round(vl + float64(t.midiTrack.pitchbendGlide.startValue))
				if vll > 8192 {
					vll = 8192
				}

				if vll < -8191 {
					vll = -8191
				}

				evts = append(evts,
					&Event{
						Position: TicksAbsPos(t.midiTrack.pitchbendGlide.startPosition) + TicksAbsPos(t.smf.writer.ticks)*TicksAbsPos(step)/16,
						Message:  midi.Pitchbend(ch, int16(vll)).Bytes(),
					})

			})
			t.midiTrack.pitchbendGlide.active = false
		}

		evts = append(evts,
			&Event{
				Position: pos,
				Message:  midi.Pitchbend(ch, v.Value).Bytes(),
			})

		if len(v.Tilde) > 0 {
			t.midiTrack.pitchbendGlide.startPosition = uint64(pos)
			t.midiTrack.pitchbendGlide.startValue = v.Value
			t.midiTrack.pitchbendGlide.active = true
			t.midiTrack.pitchbendGlide.glideFunc = items.LinearGlide
			if v.Tilde == "~~" {
				t.midiTrack.pitchbendGlide.glideFunc = items.ExponentialGlide
			}
		}

		if v.Dotted != "" && v.Value != 0 {
			l := items.DottedLengths[v.Dotted][1]
			evts = append(evts, &Event{
				Position: TicksAbsPos(uint16(pos) + (t.smf.writer.ticks * 4 / l)),
				Message:  midi.Pitchbend(ch, 0).Bytes()})
		}

	case *items.MultiItem:
		var _evs Events

		for i, it := range v.Events {
			___ev := it.Dup()
			___ev.PosShift += sketchEvent.PosShift
			___ev.Position = sketchEvent.Position

			_ev, err := t.convertSketchEvent(ch, ___ev, trackDelay)
			if err != nil {
				return nil, fmt.Errorf("error while converting multiitem#%v (%s) at position: %v: %s", i, it, pos, err)
			}

			_evs = append(_evs, _ev...)
		}

		sort.Sort(_evs)
		var didStop bool

		for _, _ev := range _evs {
			if _ev.StopNotes && didStop {
				_ev.StopNotes = false
			}
			if _ev.StopNotes {
				didStop = true
			}
			evts = append(evts, _ev)
		}

	case *items.NTuple:
		length := TicksAbsPos(v.LengthTicks)
		//fmt.Printf("ntuple length32th: %v ticks: %v\n", v.LengthTicks, length)
		plength := length / TicksAbsPos(len(v.Events))
		var delta TicksAbsPos

		for _, it := range v.Events {

			_ev := it.Dup()
			_ev.PosShift += sketchEvent.PosShift
			_ev.Position = sketchEvent.Position

			if ntpp, is := _ev.Item.(*items.NTuple); is {
				ntpp.LengthTicks = uint(plength)
				_ev.Item = ntpp
			}

			//fmt.Printf("event: %#v, delta: %v\n", _ev, delta)

			__evts, err := t.convertSketchEvent(ch, _ev, trackDelay)
			if err != nil {
				return nil, err
			}

			for _, __evt := range __evts {
				__evt.Position += delta
				evts = append(evts, __evt)
			}

			delta += plength
		}

		evts = append(evts, &Event{
			Position: pos + length,
			//message:  meta.Undefined{0, nil}, stopNotes: true})
			Message: smf.MetaUndefined(0, nil), StopNotes: true})

	case *items.MidiSample:
		//length := v.LengthTicks
		err = v.Fetch(t.smf.score.FS, items.MIDI_SAMPLES_DIR, t.smf.writer.ticks)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			return
		}

		// var totalPosition uint
		var additional TicksAbsPos

		addDelta := pos - TicksAbsPos(v.NegativeOffset)

		abs := addDelta

		var firstNote = true

		for i, tev := range v.Track {

			//		fmt.Printf("additional: %v\n", additional)

			// we change the channel everywhere to the track channel (ch)
			// therefor we scan the original channel to _ch, but ignore it
			var _ch, byte1, byte2 uint8
			var pbrel int16
			var pbabs uint16

			var msg smf.Message
			var ev = &Event{}

			switch {

			case tev.Message.GetAfterTouch(&_ch, &byte1):
				msg = smf.Message(midi.AfterTouch(ch, byte1))

			case tev.Message.GetControlChange(&_ch, &byte1, &byte2):
				msg = smf.Message(midi.ControlChange(ch, byte1, byte2))

			case tev.Message.GetNoteStart(&_ch, &byte1, &byte2):
				if firstNote {
					ev.StopNotes = true
				}
				firstNote = false
				if i == len(v.Track)-1 {
					ev.Monitor = true
				}
				msg = smf.Message(midi.NoteOn(ch, byte1, byte2))

			case tev.Message.GetNoteEnd(&_ch, &byte1):
				msg = smf.Message(midi.NoteOff(ch, byte1))

			case tev.Message.GetPolyAfterTouch(&_ch, &byte1, &byte2):
				msg = smf.Message(midi.PolyAfterTouch(ch, byte1, byte2))

			case tev.Message.GetPitchBend(&_ch, &pbrel, &pbabs):
				msg = smf.Message(midi.Pitchbend(ch, pbrel))

			case tev.Message.GetProgramChange(&_ch, &byte1):
				// ignore program changes
				additional += TicksAbsPos(tev.Delta)
				continue
			default:
				additional += TicksAbsPos(tev.Delta)
				continue
			}

			abs += TicksAbsPos(tev.Delta) + additional
			ev.Position = abs
			ev.Message = msg
			evts = append(evts, ev)
			additional = 0
		}

	case *items.GlideStart:
		t.midiTrack.noteGlide.startNote = t.midiTrack.PrevKey
		evts = append(evts,
			&Event{
				Position: pos,
				//stopNotes: true,
				Message: midi.Pitchbend(ch, 0).Bytes(),
			})

		t.midiTrack.noteGlide.startPosition = uint64(pos)
		t.midiTrack.noteGlide.active = true
		t.midiTrack.noteGlide.glideFunc = items.LinearGlide
		if *v == items.GlideExponential {
			t.midiTrack.noteGlide.glideFunc = items.ExponentialGlide
		}

	case *items.MIDIPolyAftertouch:
		ki := uint8(int8(v.Key) + t.midiTrack.Track.MIDITranspose)

		if t.midiTrack.polyaftertouchGlide.active && v.Key == t.midiTrack.polyaftertouchGlide.key {
			distance := int64(math.Round(float64(pos-TicksAbsPos(t.midiTrack.polyaftertouchGlide.startPosition)) / float64(t.midiTrack.GlideResolution(t.smf.writer.ticks))))
			diff := int64(v.Value) - int64(t.midiTrack.polyaftertouchGlide.startValue)

			t.midiTrack.polyaftertouchGlide.glideFunc(distance, diff, func(step uint, vl float64) {
				vll := math.Round(vl + float64(t.midiTrack.polyaftertouchGlide.startValue))
				if vll > 127 {
					vll = 127
				}

				if vll < 0 {
					vll = 0
				}

				evts = append(evts,
					&Event{
						Position: TicksAbsPos(uint(t.midiTrack.polyaftertouchGlide.startPosition) + uint(t.smf.writer.ticks)*step/16),
						Message:  midi.PolyAfterTouch(ch, t.midiTrack.polyaftertouchGlide.key, uint8(vll)).Bytes(),
					})
			})
			t.midiTrack.polyaftertouchGlide.active = false
		}

		evts = append(evts,
			&Event{
				Position: pos,
				Message:  midi.PolyAfterTouch(ch, ki, v.Value).Bytes(),
			})

		if len(v.Tilde) > 0 {
			t.midiTrack.polyaftertouchGlide.key = ki
			t.midiTrack.polyaftertouchGlide.startPosition = uint64(pos)
			t.midiTrack.polyaftertouchGlide.startValue = v.Value
			t.midiTrack.polyaftertouchGlide.active = true
			t.midiTrack.polyaftertouchGlide.glideFunc = items.LinearGlide
			if v.Tilde == "~~" {
				t.midiTrack.polyaftertouchGlide.glideFunc = items.ExponentialGlide
			}
		}

		if v.Dotted != "" && v.Value != 0 {
			l := items.DottedLengths[v.Dotted][1]
			evts = append(evts, &Event{
				Position: TicksAbsPos(uint(uint16(pos) + (t.smf.writer.ticks * 4 / l))),
				Message:  midi.PolyAfterTouch(ch, ki, 0).Bytes()})
		}
	default:
		err = fmt.Errorf("sketch event of type %T is not supported in smf package", sketchEvent.Item)
	}

	return
}

func (t *Track) findNextNonEmptyPos(sketchEvents []*items.Event) TicksAbsPos {
	for _, ev := range sketchEvents {
		if ev != nil && ev.Item != nil && ev.Item != items.Hold {
			return t.smf.posToTicks(ev.Position)
		}
	}
	return 0
}

func (iw *Track) trackScoreEvents(tr *track.Track, sketchEvents []*items.Event) (evts Events, err error) {
	//fmt.Printf("trackScoreEvents for track %q\n", tr.Name)

	iw.midiTrack.Reset()

	iw.midiTrack.Track.MIDITranspose = tr.MIDITranspose
	iw.midiTrack.Track.PitchbendRange = tr.PitchbendRange
	iw.midiTrack.VelocityScale = tr.VelocityScale

	var trackDelay int
	if tr.Delay[0] != 0 && tr.Delay[1] != 0 {
		trackDelay = (int(iw.smf.writer.ticks) * 4 * tr.Delay[0] / tr.Delay[1])
	}
	//fmt.Printf("trackDelay %v\n", trackDelay)

	for i, ev := range sketchEvents {
		pos := TicksAbsPos(ev.AbsPosTicks(iw.smf.writer.ticks, trackDelay))
		//fmt.Printf("trackScoreEvents, ticks: %v pos: %v\n", iw.smf.writer.ticks, pos)

		if ev.Item != nil {
			if ntp, is := ev.Item.(*items.NTuple); is {
				var nextPos TicksAbsPos
				if i < len(sketchEvents)+1 {
					nextPos = iw.findNextNonEmptyPos(sketchEvents[i+1:])
					//fmt.Printf("found findNextNonEmptyPos: %v\n", nextPos)
					if nextPos == 0 {
						nextPos = iw.smf.posToTicks(tr.EndPos)
					}
				} else {
					nextPos = iw.smf.posToTicks(tr.EndPos)
				}
				_nextPos := int(nextPos) + trackDelay
				if _nextPos < 0 {
					_nextPos = 0
				}
				nextPos = TicksAbsPos(_nextPos)
				//fmt.Printf("nextPos %v - pos %v\n", nextPos, pos)

				ntp.LengthTicks = uint(nextPos - pos)
				ev.Item = ntp
			}

			if msp, is := ev.Item.(*items.MidiSample); is {
				var nextPos TicksAbsPos
				if i < len(sketchEvents)+1 {
					nextPos = iw.findNextNonEmptyPos(sketchEvents[i+1:])
					//fmt.Printf("found findNextNonEmptyPos: %v\n", nextPos)
					if nextPos == 0 {
						nextPos = iw.smf.posToTicks(tr.EndPos)
					}
				} else {
					nextPos = iw.smf.posToTicks(tr.EndPos)
				}
				_nextPos := int(nextPos) + trackDelay
				if _nextPos < 0 {
					_nextPos = 0
				}
				nextPos = TicksAbsPos(_nextPos)
				//fmt.Printf("nextPos %v - pos %v\n", nextPos, pos)

				msp.LengthTicks = uint(nextPos - pos)
				ev.Item = msp
			}

		}

		var evnew Events
		evnew, err = iw.convertSketchEvent(uint8(tr.MIDIChannel), ev, trackDelay)
		if err != nil {
			return
		}
		evts = append(evts, evnew...)
	}
	return
}
