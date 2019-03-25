package items

import (
	"fmt"
	"strings"
)

// TODO translate augmentation symbols like ^ and ° to pitchbend messages

type Note struct {
	Letter         string
	Octave         int
	Augmenter      string
	Velocity       int8
	GlissandoStart bool
	GlissandoExp   bool
	Dotted         bool
	ScaleNote      int8
	PosShift       int // 0 = no, 1 = laidback, -1 = ahead of time
}

func (note Note) ToMIDI() (midinote_ uint8) {
	midinote := 48 // c

	if note.Octave > 0 {
		midinote += 12 * (note.Octave - 1)
	}

	if note.Octave < 0 {
		midinote += 12 * note.Octave
	}

	switch strings.ToLower(note.Letter) {
	case "c":
		midinote += 0
	case "d":
		midinote += 2
	case "e":
		midinote += 4
	case "f":
		midinote += 5
	case "g":
		midinote += 7
	case "a":
		midinote += 9
	case "b":
		midinote += 11
	}

	switch note.Augmenter {
	case "":
	case "#":
		midinote += 1
	case "^":
	case "°":
	}

	if midinote > 127 {
		midinote = 127
	}

	if midinote < 0 {
		midinote = 0
	}

	return uint8(midinote)

}

func KeyToNote(key uint8) (letter, augmenter string, octave int) {
	nt := key % 12
	octave = int(key) / 12
	//midinote := 48 // c
	if octave >= 4 {
		octave -= 3
	} else {
		octave -= 4
	}

	// octave -= 3
	if octave > 10 {
		octave = 10
	}

	switch nt {
	case 0:
		letter = "c"
	case 1:
		letter = "c"
		augmenter = "#"
	case 2:
		letter = "d"
	case 3:
		letter = "d"
		augmenter = "#"
	case 4:
		letter = "e"
	case 5:
		letter = "f"
	case 6:
		letter = "f"
		augmenter = "#"
	case 7:
		letter = "g"
	case 8:
		letter = "g"
		augmenter = "#"
	case 9:
		letter = "a"
	case 10:
		letter = "a"
		augmenter = "#"
	case 11:
		letter = "b"
	}
	return
}

func (n Note) String() string {
	var bf strings.Builder

	if n.ScaleNote == 0 {
		letter := n.Letter
		if n.Octave < 0 {
			letter = strings.ToUpper(letter)
		}

		bf.WriteString(letter)
		bf.WriteString(n.Augmenter)

		switch n.Octave {
		case -5, 5:
			bf.WriteString("\"\"")
		case -4, 4:
			bf.WriteString("\"'")
		case -3, 3:
			bf.WriteString("\"")
		case -2, 2:
			bf.WriteString("'")
		case 1:
			// do nothing
		case 0:
			panic("invalid octave")
		case -1:
			// do nothing
		case 6:
			bf.WriteString("\"\"'")
		case 7:
			bf.WriteString("\"\"\"")
		case 8:
			bf.WriteString("\"\"\"'")
		case 9:
			bf.WriteString("\"\"\"\"")
		case 10:
			bf.WriteString("\"\"\"\"'")
		// case 10:
		// bf.WriteString("\"\"\"\"\"")
		default:
			panic(fmt.Sprintf("invalid octave: %v", n.Octave))
		}
	} else {
		bf.WriteString(fmt.Sprintf("^%v", n.ScaleNote))
	}

	bf.WriteString(velocityToDynamic(n.Velocity))

	if n.Dotted {
		bf.WriteString(":")
	}

	if n.GlissandoStart {
		bf.WriteString("~")
	}

	if n.GlissandoExp {
		bf.WriteString("~")
	}

	if n.PosShift == 1 {
		bf.WriteString(">")
	}

	if n.PosShift == -1 {
		bf.WriteString("<")
	}

	return bf.String()
}

func (n Note) Dup() Note {
	return n
}

type MIDINote struct {
	Note     int8
	Velocity int8
	Dotted   bool
	PosShift int // 0 = no, 1 = laidback, -1 = ahead of time
}

func (m MIDINote) String() string {
	vel := velocityToDynamic(m.Velocity)
	dotted := ""
	if m.Dotted {
		dotted = ":"
	}

	posShift := ""

	if m.PosShift == 1 {
		posShift = ">"
	}

	if m.PosShift == -1 {
		posShift = "<"
	}

	return fmt.Sprintf("MN%v%s%s%s", m.Note, vel, dotted, posShift)
}

func (mn MIDINote) Dup() MIDINote {
	return mn
}

type MIDICC struct {
	Controller uint8
	Value      uint8
	Dotted     bool
	Tilde      string
}
type MIDIPitchbend struct {
	Value  int16
	Dotted bool
	Tilde  string
}
type MIDIPolyAftertouch struct {
	Key    uint8
	Value  uint8
	Dotted bool
	Tilde  string
}
type MIDIAftertouch struct {
	Value  uint8
	Dotted bool
	Tilde  string
}

type OSCMessage struct {
	Path string
	Args []interface{}
}

type NTuple struct {
	EndPos   uint // 32th in bar from the beginning
	Items    []interface{}
	PosShift int // 0 = no, 1 = laidback, -1 = ahead of time
}
type RepeatLastEvent struct{}
type RepeatLastBar struct{}
type RepeatLastBarUntilChange struct{}
type RepeatLastNBarsUntilChange int
type rest struct{}
type hold struct{}
type GlissStart bool
type Lyric string

type RandomProbability struct {
	Prob             uint8
	Item             interface{}
	Chosen           bool
	ItemOriginalData string
}

type RandomChooser struct {
	Alternatives             []interface{}
	AlternativesOriginalData []string
	Chosen                   int
}

var Hold = hold{}
var Rest = rest{}
var GlissandoLinear = GlissStart(false)
var GlissandoExponential = GlissStart(true)

type MultiItem []interface{}
