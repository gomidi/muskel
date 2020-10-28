package track

import (
	"fmt"
	"strconv"
	"strings"
)

type Track struct {
	Name           string
	FileGroup      string
	Import         string
	MIDIChannel    int8
	MIDIProgram    int8
	MIDIBank       int8
	MIDIVolume     int8
	MIDITranspose  int8   // only transposes Notes not MIDINotes
	Delay          [2]int // fraction of 4/4
	PitchbendRange uint8
	// velocity-scale property to refine velocity conversion (min, max, randomize-factor, step-width)
	VelocityScale [5]uint8 // 0: min 1: max 2: randomize-factor 3: step-width, 4: center
	/*
		ambitus from: to:
	*/
	Ambitus            [2]string // notes
	AmbitusCutOverFlow bool
	EndPos             uint
}

func New(name string) *Track {
	var tr = &Track{}
	tr.Delay = [2]int{0, 0}
	tr.MIDIChannel = -1
	tr.MIDIProgram = -1
	tr.MIDIVolume = -1
	tr.MIDIBank = -1
	tr.MIDITranspose = 0
	tr.PitchbendRange = 2
	tr.VelocityScale = DefaultVeloctiyScale
	tr.FileGroup = ""
	tr.Import = ""
	tr.Name = name
	return tr
}

func (t *Track) SetImport(importFrom string) error {
	t.Import = importFrom
	return nil
}

var DefaultVeloctiyScale = [5]uint8{1, 127, 4, 15, 63}

func (t *Track) SetAmbitus(ambString string) error {
	if strings.TrimSpace(ambString) == "" {
		return nil
	}
	props := strings.Split(ambString, " ")

	for _, prop := range props {
		prop = strings.TrimSpace(prop)

		if prop == "" {
			continue
		}
		pr := strings.Split(prop, ":")

		if len(pr) != 2 {
			return fmt.Errorf("error in Ambitus property of track %q: %q", t.Name, ambString)
		}

		key := strings.TrimSpace(pr[0])

		switch strings.ToLower(key) {
		case "from":
			t.Ambitus[0] = strings.TrimSpace(pr[1])
		case "to":
			t.Ambitus[1] = strings.TrimSpace(pr[1])
		case "cut":
			val, errConv := strconv.ParseBool(strings.TrimSpace(pr[1]))

			if errConv != nil {
				return fmt.Errorf("error in Ambitus property of track %q: %q: %s is not a bool", t.Name, ambString, strings.TrimSpace(pr[1]))
			}
			t.AmbitusCutOverFlow = val
		default:
			return fmt.Errorf("error in Ambitus property of track %q: unknown property %q", t.Name, key)
		}
	}

	return nil
}

func (t *Track) SetVelScale(velScale string) error {
	//  VelocityScale  | min:1 max:127 random:4 step:15 center:63 |
	// "min: %v max: %v random: %v step: %v"
	//0: min 1: max 2: randomize-factor 3: step-width, 4: center

	//fmt.Printf("velScale of track %s: %q\n", track.Name, velScale)

	props := strings.Split(velScale, " ")

	//fmt.Printf("props: %#v\n", props)

	res := DefaultVeloctiyScale

	for _, prop := range props {
		prop = strings.TrimSpace(prop)

		if prop == "" {
			continue
		}
		pr := strings.Split(prop, ":")

		if len(pr) != 2 {
			return fmt.Errorf("error in VelScale property of track %q: %q", t.Name, velScale)
		}

		key := strings.TrimSpace(pr[0])

		val, errConv := strconv.Atoi(strings.TrimSpace(pr[1]))

		if errConv != nil {
			return fmt.Errorf("error in VelScale property of track %q: %q: %s is not an int", t.Name, velScale, strings.TrimSpace(pr[1]))
		}

		switch strings.ToLower(key) {
		case "min":
			res[0] = uint8(val)
		case "max":
			res[1] = uint8(val)
		case "random":
			res[2] = uint8(val)
		case "step":
			res[3] = uint8(val)
		case "center":
			res[4] = uint8(val)
		default:
			return fmt.Errorf("error in VelScale property of track %q: unknown property %q", t.Name, key)
		}
	}

	//fmt.Printf("velocity scale set to %v for track %s\n", res, track.Name)

	t.VelocityScale = res

	return nil
}

func (t *Track) SetPitchbendrange(data string) error {
	if len(data) == 0 {
		return nil
	}
	rg, errC := strconv.Atoi(data)
	if errC != nil {
		return fmt.Errorf("%#v is not a valid MIDI pitchbendrange", data)
	}
	if rg > 24 || rg < 1 {
		return fmt.Errorf("%#v is not a valid MIDI pitchbendrange", rg)
	}

	t.PitchbendRange = uint8(rg)
	return nil
}

// setTrackChannel sets the MIDI channel of a track
func (t *Track) SetChannel(data string) error {
	if len(data) == 0 {
		t.MIDIChannel = -1
		return nil
	}
	cc, errC := strconv.Atoi(data)
	if errC != nil {
		return fmt.Errorf("%#v is not a valid MIDI channel", data)
	}
	if cc > 16 || cc < 1 {
		return fmt.Errorf("%#v is not a valid MIDI channel", cc)
	}

	ch := int8(cc - 1)

	if t.MIDIChannel == -1 && data != "" {
		t.MIDIChannel = ch
		return nil
	}

	if ch != t.MIDIChannel {
		return fmt.Errorf("can't set MIDI channel for track %q to %v: MIDI channel already set to: %v", t.Name, ch, t.MIDIChannel)
	}

	return nil
}

func (t *Track) SetFileGroup(data string) error {
	t.FileGroup = data
	return nil
}

// setTrackProgram sets the MIDI program value of a track
func (t *Track) SetProgram(data string) error {
	if len(data) == 0 {
		t.MIDIProgram = -1
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI program value", data)
	}
	if pr > 128 || pr < 1 {
		return fmt.Errorf("%#v is not a valid MIDI program value", pr)
	}

	prog := int8(pr - 1)

	if t.MIDIProgram == -1 && data != "" {
		t.MIDIProgram = prog
		return nil
	}

	if prog != t.MIDIProgram {
		return fmt.Errorf("can't set MIDI program for track %q to %v: MIDI program already set to: %v", t.Name, prog, t.MIDIProgram)
	}

	return nil
}

// setTrackVolume sets the MIDI volume of a track
func (t *Track) SetVolume(data string) error {
	if len(data) == 0 {
		t.MIDIVolume = -1
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI volume value", data)
	}
	if pr > 128 || pr < 1 {
		return fmt.Errorf("%#v is not a valid MIDI volume value", pr)
	}

	vol := int8(pr - 1)

	if t.MIDIVolume == -1 && data != "" {
		t.MIDIVolume = vol
		return nil
	}

	if vol != t.MIDIVolume {
		return fmt.Errorf("can't set MIDI volume for track %q to %v: MIDI volume already set to: %v", t.Name, vol, t.MIDIVolume)
	}

	return nil
}

// setTrackTranspose sets the MIDITranspose for a track
func (t *Track) SetTranspose(data string) error {
	if len(data) == 0 || data == "" {
		t.MIDITranspose = 0
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI transpose value", data)
	}
	if pr >= 84 || pr <= -84 {
		return fmt.Errorf("%#v is not a valid MIDI transpose value (must be >=-84 and <=84", pr)
	}

	tr := int8(pr)

	t.MIDITranspose = tr
	return nil
}

// setTrackBank sets the MIDI bank for a track
func (t *Track) SetBank(data string) error {
	if len(data) == 0 {
		t.MIDIBank = -1
		return nil
	}
	pr, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("%#v is not a valid MIDI bank value", data)
	}
	if pr > 128 || pr < 1 {
		return fmt.Errorf("%#v is not a valid MIDI bank value", pr)
	}

	bank := int8(pr - 1)

	if t.MIDIBank == -1 && data != "" {
		t.MIDIBank = bank
		return nil
	}

	if bank != t.MIDIBank {
		return fmt.Errorf("can't set MIDI bank for track %q to %v: MIDI bank already set to: %v", t.Name, bank, t.MIDIBank)
	}

	return nil
}

// setDelay sets the delay for a track
func (t *Track) SetDelay(data string) error {
	if len(data) == 0 {
		t.Delay = [2]int{0, 0}
		return nil
	}

	d := strings.Split(data, "/")

	if len(d) != 2 {
		return fmt.Errorf("%#v is not a valid delay value", data)
	}

	num, err := strconv.Atoi(d[0])
	if err != nil {
		return fmt.Errorf("%#v is not a valid delay numerator", d[0])
	}

	denom, err := strconv.Atoi(d[1])
	if err != nil || denom == 0 {
		return fmt.Errorf("%#v is not a valid delay denominator", d[1])
	}

	t.Delay = [2]int{num, denom}
	//fmt.Printf("setting delay for instrument %q to %v/%v\n", instr.Name, num, denom)

	return nil
}
