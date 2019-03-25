package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel/muskellib"
)

var regexMIDINote = regexp.MustCompile("^([0-9]{1,3})([-+" + regexp.QuoteMeta("=") + "]*)(:{0,1})([<>]{0,1})$")

func parseMIDINote(data string) (nt muskellib.MIDINote, err error) {
	mt := regexMIDINote.FindStringSubmatch(data)
	if len(mt) < 2 {
		return nt, fmt.Errorf("not a valid MIDINote: %q", data)
	}
	//	fmt.Printf("parse MIDInotes: %#v\n", mt)
	i, err := strconv.Atoi(mt[1])
	if err != nil {
		return nt, fmt.Errorf("not a valid MIDINote: %q: %s", data, err.Error())
	}
	if i < 1 && i > 128 {
		return nt, fmt.Errorf("not a valid MIDINote: %q: must be >= 1 and <= 128", data)
	}
	nt.Note = int8(i)
	nt.Velocity = muskellib.DynamicToVelocity(mt[2])
	nt.Dotted = mt[3] == ":"

	if mt[4] == ">" {
		nt.PosShift = 1
	}

	if mt[4] == "<" {
		nt.PosShift = -1
	}
	return nt, nil
}

func parseMIDICC(data string) (cc muskellib.MIDICC, err error) {
	var ccval int = -1

	if len(data) > 2 {
		if strings.ToLower(data[0:2]) == "on" {
			ccval = 127
			data = data[2:]
		}

		if strings.ToLower(data[0:3]) == "off" {
			ccval = 0
			data = data[3:]
		}
	}

	if idx := strings.Index(data, "~"); idx > -1 {
		cc.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		cc.Dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	d := strings.Split(dt, ",")
	if ccval == -1 && len(d) != 2 {
		err = fmt.Errorf("invalid format for MIDI CC: %#v, must be CC(nn,mm) where nn is the controller number and mm is the controler value", data)
		return
	}

	var ccnum int
	ccnum, err = strconv.Atoi(d[0])

	if err != nil {
		return
	}

	if ccval > -1 {
		cc.Controller = uint8(ccnum)
		cc.Value = uint8(ccval)
		return
	}

	ccval, err = strconv.Atoi(d[1])

	if err != nil {
		return
	}

	cc.Controller = uint8(ccnum)
	cc.Value = uint8(ccval)

	return
}

func parseMIDIPitchbend(data string) (pb muskellib.MIDIPitchbend, err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		pb.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		pb.Dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	// fmt.Printf("parsing pitchbend %q: %v, normalized: %v\n", dt, val, int16(val))

	pb.Value = int16(val)
	return
}

func parseMIDIAftertouch(data string) (at muskellib.MIDIAftertouch, err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		at.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		at.Dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	var val int
	val, err = strconv.Atoi(dt)

	if err != nil {
		return
	}

	at.Value = uint8(val)
	return
}

func parseMIDIPolyAftertouch(data string) (pt muskellib.MIDIPolyAftertouch, err error) {
	if idx := strings.Index(data, "~"); idx > -1 {
		pt.Tilde = data[idx:]
		data = data[:idx]
	}

	if idx := strings.Index(data, ":"); idx > -1 {
		pt.Dotted = true
		data = data[:idx]
	}

	dt := strings.Trim(data, "()")
	d := strings.Split(dt, ",")
	if len(d) != 2 {
		err = fmt.Errorf("invalid format for MIDI PT: %#v, must be PT(nt,mm) where nt is the note and mm is the value", data)
		return
	}

	var nt muskellib.Note
	nt, err = parseNote(d[0])
	if err != nil {
		err = fmt.Errorf("invalid format for MIDI PT can't parse note: %s", err.Error())
		return
	}

	var val int
	val, err = strconv.Atoi(d[1])

	if err != nil {
		err = fmt.Errorf("invalid format for MIDI PT can't parse value: %s", err.Error())
		return
	}

	pt.Key = nt.ToMIDI()
	pt.Value = uint8(val)

	return
}

var regScaleNote = regexp.MustCompile("^(-){0,1}([0-9]+)(.*)$")

func parseScaleNote(data string) (nt muskellib.Note, err error) {
	mt := regScaleNote.FindStringSubmatch(data)
	// fmt.Printf("scale note %q match: %#v\n", data, mt)
	var i int
	i, err = strconv.Atoi(mt[1] + mt[2])

	if err != nil {
		err = fmt.Errorf("invalid scale note: %#v", data)
		return
	}

	nt.ScaleNote = int8(i)

	var dynamic string

	for _, l := range mt[3] {
		switch l {
		case ':':
			nt.Dotted = true
		case '+':
			dynamic += "+"
		case '-':
			dynamic += "-"
		case '=':
			dynamic = "="
		case '~':
			nt.GlissandoStart = true
		case '#':
			nt.Augmenter += "#"
		case '^':
			nt.Augmenter += "^"
		case '°':
			nt.Augmenter += "°"
		case '>':
			nt.PosShift = 1
		case '<':
			nt.PosShift = -1
		/*
			case '#':
				nt.augmenter = "#"
			case '^':
				nt.augmenter = "^"
			case '°':
				nt.augmenter = "°"
			case '"':
				if nt.octave > 0 {
					nt.octave += 2
				}

				if nt.octave < 0 {
					nt.octave -= 2
				}
			case '\'':
				if nt.octave > 0 {
					nt.octave += 1
				}

				if nt.octave < 0 {
					nt.octave -= 1
				}
		*/
		default:
			err = fmt.Errorf("invalid scale note: %#v", data)
			return
		}
	}

	nt.Velocity = muskellib.DynamicToVelocity(dynamic)

	return
}

func parseNote(data string) (nt muskellib.Note, err error) {

	switch data[:1] {
	case "a", "b", "c", "d", "e", "f", "g":
		nt.Letter = data[:1]
		nt.Octave = 1
	case "A", "B", "C", "D", "E", "F", "G":
		nt.Letter = strings.ToLower(data[:1])
		nt.Octave = -1
	default:
		err = fmt.Errorf("invalid note: %#v", data)
		return
	}

	var dynamic string
	var gliss int

	for _, l := range data[1:] {
		switch l {
		case ':':
			nt.Dotted = true
		case '+':
			dynamic += "+"
		case '-':
			dynamic += "-"
		case '=':
			dynamic = "="
		case '~':
			gliss++
		case '>':
			nt.PosShift = 1
		case '<':
			nt.PosShift = -1
		case '#':
			nt.Augmenter += "#"
		case '^':
			nt.Augmenter += "^"
		case '°':
			nt.Augmenter += "°"
		case '"':
			if nt.Octave > 0 {
				nt.Octave += 2
			}

			if nt.Octave < 0 {
				nt.Octave -= 2
			}
		case '\'':
			if nt.Octave > 0 {
				nt.Octave += 1
			}

			if nt.Octave < 0 {
				nt.Octave -= 1
			}
		default:
			err = fmt.Errorf("invalid note: %#v", data)
			return
		}
	}

	if gliss > 0 {
		nt.GlissandoStart = true
	}

	if gliss > 1 {
		nt.GlissandoExp = true
	}

	nt.Velocity = muskellib.DynamicToVelocity(dynamic)

	return
}
