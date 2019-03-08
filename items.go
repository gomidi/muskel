package muskel

import (
	"bytes"
	"fmt"
	"strings"
)

// TODO see the following list

/*
+ accel. decel. cresc. descr. (wahrscheinlich mit Volume CC)
- temperaments
- ggf. Chords
- translate augmentation symbols like ^ and ° to pitchbend messages
*/

type Note struct {
	letter         string
	octave         int
	augmenter      string
	velocity       int8
	glissandoStart bool
	dotted         bool
	scaleNote      int8
}

func (note Note) toMIDI() (midinote_ uint8) {
	midinote := 48 // c

	if note.octave > 0 {
		midinote += 12 * (note.octave - 1)
	}

	if note.octave < 0 {
		midinote += 12 * note.octave
	}

	switch strings.ToLower(note.letter) {
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

	switch note.augmenter {
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

func keyToNote(key uint8) (letter, augmenter string, octave int) {
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
	var bf bytes.Buffer

	if n.scaleNote == 0 {
		letter := n.letter
		if n.octave < 0 {
			letter = strings.ToUpper(letter)
		}

		bf.WriteString(letter)
		bf.WriteString(n.augmenter)

		switch n.octave {
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
			panic(fmt.Sprintf("invalid octave: %v", n.octave))
		}
	} else {
		bf.WriteString(fmt.Sprintf("^%v", n.scaleNote))
	}

	bf.WriteString(velocityToDynamic(n.velocity))

	if n.dotted {
		bf.WriteString(":")
	}

	if n.glissandoStart {
		bf.WriteString("~")
	}

	return bf.String()
}

func (n Note) Dup() Note {
	return n
}

type MIDINote struct {
	note     int8
	velocity int8
	dotted   bool
}

func (mn MIDINote) Dup() MIDINote {
	return mn
}

type MIDICC [2]uint8
type MIDIPitchbend int16
type MIDIPolyAftertouch [2]uint8
type MIDIAftertouch uint8

type OSCMessage struct {
	Path string
	Args []interface{}
}

type NTuple struct {
	endPos uint // 32th in bar from the beginning
	items  []interface{}
}
type RepeatLastEvent struct{}
type RepeatLastBar struct{}
type RepeatLastBarUntilChange struct{}
type RepeatLastNBarsUntilChange int
type rest struct{}
type hold struct{}
type Lyric string

type RandomProbability struct {
	prob             uint8
	item             interface{}
	chosen           bool
	itemOriginalData string
}

type RandomChooser struct {
	alternatives             []interface{}
	alternativesOriginalData []string
	chosen                   int
}

var Hold = hold{}
var Rest = rest{}

/*
- notes, e.g. a'
- drumnotes e.g. sn
- motivs e.g. $a
- MIDI controller events e.g. PB200
- rests e.g. _
- samples with a timespan e.g. Sa
- samples with no timespan ("oneshot") e.g. Za
- repitition of last event e.g. %
- multiitems, consisting of several items happening at the same time. e.g. (a',PB200)
*/

type MultiItem []interface{}
