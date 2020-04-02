package items

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Note struct {
	Letter         string
	Octave         int
	Augmenter      string
	Dynamic        string
	GlissandoStart bool
	GlissandoExp   bool
	Dotted         string
	IsScaleNote    bool
	ScaleNote      int8
	PosShift       int // 0 = no, 1 = laidback, -1 = ahead of time
	NoteOn         bool
	NoteOff        bool
}

var _ Item = &Note{}

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
	//case "^":
	//case "°":
	case "b":
		midinote -= 1
		//case "^":
		//case "°":
	}

	//midinote += int(note.Transposition)

	if midinote > 127 {
		midinote = 127
	}

	if midinote < 0 {
		midinote = 0
	}

	return uint8(midinote)

}

func (n Note) String() string {
	var bf strings.Builder

	if !n.IsScaleNote {
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
		if n.ScaleNote >= 0 {
			bf.WriteString(fmt.Sprintf("%v", n.ScaleNote+1))
		} else {
			bf.WriteString(fmt.Sprintf("%v", n.ScaleNote))
		}

		switch n.Octave {
		case -4, 4:
			bf.WriteString("\"\"")
		case -3, 3:
			bf.WriteString("\"'")
		case -2, 2:
			bf.WriteString("\"")
		case -1, 1:
			bf.WriteString("'")
		case 0:
			// do nothing
		case -5, 5:
			bf.WriteString("\"\"'")
		case -6, 6:
			bf.WriteString("\"\"\"")
		case -7, 7:
			bf.WriteString("\"\"\"'")
		case -8, 8:
			bf.WriteString("\"\"\"\"")
		case -9, 9:
			bf.WriteString("\"\"\"\"'")
		default:
			panic(fmt.Sprintf("invalid octave: %v", n.Octave))
		}
	}

	//bf.WriteString(velocityToDynamic(n.Velocity))
	bf.WriteString(n.Dynamic)

	if n.Dotted != "" {
		bf.WriteString(n.Dotted)
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

	if n.NoteOn {
		return bf.String() + "_"
	}

	if n.NoteOff {
		return "_" + bf.String()
	}

	return bf.String()
}

func (n Note) Dup() Item {
	return &n
}

func parseTransposition(data string) (trans int8, rest string, err error) {
	var numbf bytes.Buffer
	var restbf bytes.Buffer
	var numComplete bool
	var minusPossible = true

	for _, char := range data {
		switch char {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			minusPossible = false
			if !numComplete {
				numbf.WriteRune(char)
			} else {
				restbf.WriteRune(char)
			}
		case '-':
			if !minusPossible {
				numComplete = true
			}
			if !numComplete {
				numbf.WriteRune(char)
			} else {
				restbf.WriteRune(char)
			}
		default:
			if !numComplete {
				numComplete = true
			}
			restbf.WriteRune(char)
		}
	}

	tr, err := strconv.Atoi(numbf.String())
	if err != nil {
		return 0, "", fmt.Errorf("can't parse number of transposition in %q", data)
	}
	return int8(tr), restbf.String(), nil
}

func (nt *Note) Parse(data string, posIn32th uint) (err error) {

	switch data[0] {
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return nt.parseScale(data)
	}

	if regExEndScaleNote.MatchString(data) {
		return nt.parseScale(data)
	}

	original := data

	nt.NoteOn, nt.NoteOff, data = stripNoteOnOff(data)

	switch data[:1] {
	case "a", "b", "c", "d", "e", "f", "g":
		nt.Letter = data[:1]
		nt.Octave = 1
	case "A", "B", "C", "D", "E", "F", "G":
		nt.Letter = strings.ToLower(data[:1])
		nt.Octave = -1
	default:
		err = fmt.Errorf("invalid note: %#v I", original)
		return
	}

	if len(data) > 1 {
		data = data[1:]
	} else {
		data = ""
	}

	if len(data) > 0 && data[0] == '#' {
		nt.Augmenter = "#"
		if len(data) > 1 {
			data = data[1:]
		} else {
			data = ""
		}
	}

	if len(data) > 0 && data[0] == 'b' {
		nt.Augmenter = "b"
		if len(data) > 1 {
			data = data[1:]
		} else {
			data = ""
		}
	}

	//fmt.Printf("note data: %q\n", data)

	var dynamic string
	var gliss int

	for _, l := range data {
		switch l {
		case ':':
			nt.Dotted += ":"
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
		//case '#':
		//	nt.Augmenter += "#"
		//case '^':
		//nt.Augmenter += "^"
		//case '°':
		//nt.Augmenter += "°"
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
			err = fmt.Errorf("invalid note: %#v IV", original)
			return
		}
	}

	if gliss > 0 {
		nt.GlissandoStart = true
	}

	if gliss > 1 {
		nt.GlissandoExp = true
	}

	//nt.Velocity = DynamicToVelocity(dynamic)
	nt.Dynamic = dynamic

	return
}

var regScaleNote = regexp.MustCompile("(-){0,1}([1-9][0-9]*)(.*)$")

func (nt *Note) parseScale(data string) (err error) {
	original := data

	nt.NoteOn, nt.NoteOff, data = stripNoteOnOff(data)

	if !regScaleNote.MatchString(data) {
		err = fmt.Errorf("invalid scale note: %#v: %s", data, err)
		return
	}

	var transp int8

	if idx := strings.Index(data, "^"); idx > 0 {
		if len(data[idx:]) < 2 {
			err = fmt.Errorf("invalid scale note: %#v I", original)
		}
		pre := data[:idx]
		transp, data, err = parseTransposition(data[idx+1:])
		if err != nil {
			err = fmt.Errorf("invalid scale note: %#v: %s II", original, err)
			return
		}

		data = pre + data
	}

	mt := regScaleNote.FindStringSubmatch(data)
	//fmt.Printf("scale note %q match: %#v\n", data, mt)
	var i int
	i, err = strconv.Atoi(mt[1] + mt[2])

	if err != nil {
		err = fmt.Errorf("invalid scale note: %#v", data)
		return
	}

	nt.ScaleNote = int8(i)

	if nt.ScaleNote > 0 {
		nt.ScaleNote -= 1
	}

	nt.IsScaleNote = true
	switch {
	case transp == 0:
	default:
		nt.ScaleNote = nt.ScaleNote + transp
	}

	var dynamic string

	for _, l := range mt[3] {
		switch l {
		case ':':
			nt.Dotted += ":"
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
		case 'b':
			nt.Augmenter += "b"
		case '>':
			nt.PosShift += 1
		case '<':
			nt.PosShift -= 1
		case '"':
			if nt.ScaleNote >= 0 {
				nt.Octave += 2
			} else {
				nt.Octave -= 2
			}
		case '\'':
			if nt.ScaleNote >= 0 {
				nt.Octave += 1
			} else {
				nt.Octave -= 1
			}

		default:
			err = fmt.Errorf("invalid scale note: %#v IV", data)
			return
		}
	}

	nt.Dynamic = dynamic
	return
}
