package muskel

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// pad pads the given string with empty space to the right if it is less then length
func pad(s string, length int) string {
	if len(s) >= length {
		return s
	}

	diff := length - len(s)
	var bf bytes.Buffer

	bf.WriteString(s)

	for i := 0; i < diff; i++ {
		bf.WriteString(" ")
	}
	return bf.String()
}

// if returned vel is < 0, then it has not been set and
// is drived from the previous velocity
// always return such that rand.Intn(4) can be added or substracted
func velocityFromDynamic(dyn string) (vel int8) {
	vel = -1
	for _, rn := range dyn {
		switch rn {
		case '+':
			if vel == -1 {
				vel = 63
			}
			vel += 15
		case '-':
			if vel == -1 {
				vel = 63
			}
			vel -= 15
		case '*': // reset
			vel = 63
		}
	}

	if vel == -1 {
		return vel
	}

	if vel > 127 {
		vel = 123
	}

	if vel < 5 {
		vel = 5
	}

	return
}

func noteToMIDI(note Note) (midinote_ uint8) {
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

	return uint8(midinote)

}

type patternFragment struct {
	position string
	item     string
}

func (f *patternFragment) parse(s string) {
	if regPos.MatchString(s) {
		all := regPos.FindAllString(s, 1)
		f.position = all[0]

		s = regPos.ReplaceAllString(s, "")
	}
	f.item = s
}

// syntax for params:  #1 #2 etc.
var pattReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([1-9]+)")
var regPos = regexp.MustCompile("^([1-9]?[0-9]?)([&;" + regexp.QuoteMeta(".") + "]*)")

type positionedEvent struct {
	position     string
	originalData string
	//barNo           int
	positionIn32ths uint
	item            interface{}
}

func (p *positionedEvent) dup() (nu *positionedEvent) {
	return &positionedEvent{
		position:        p.position,
		originalData:    p.originalData,
		positionIn32ths: p.positionIn32ths,
		item:            p.item,
	}
}

func (p *positionedEvent) String() string {
	//return p.position + p.originalData
	return pos32thToString(p.positionIn32ths) + p.originalData
}

func splitItems(def string) (items []string) {
	def = strings.TrimSpace(def)

	its := strings.Split(def, " ")

	for _, it := range its {
		it = strings.TrimSpace(it)
		if it != "" {
			items = append(items, it)
		}
	}
	return
}

func replaceItemWith(replacement string) (position string, item string) {
	var frepl patternFragment
	frepl.parse(replacement)

	position = frepl.position

	item = frepl.item

	return
}

func pos32thToString(pos uint) string {
	qn := pos / 8
	rest := pos % 8

	var bf bytes.Buffer
	fmt.Fprintf(&bf, "%v", qn+1)

	eights := rest / 4
	rest = rest % 4

	if eights > 0 {
		fmt.Fprint(&bf, "&")
	}

	sixteenth := rest / 2
	rest = rest % 2

	if sixteenth > 0 {
		fmt.Fprint(&bf, ".")
	}

	if rest > 0 {
		fmt.Fprint(&bf, ";")
	}

	return bf.String()
}

// length returns the bar length in 32th
func length32ths(num, denom uint8) int {
	return int(num*32) / int(denom)
}

func getQNNumberFromPos(pos string) (qnnumber int, rest string) {
	if len(pos) == 0 {
		panic("empty position is not valid")
	}
	qnnumber = -1
	rest = pos

	if len(pos) > 1 {
		if i2, err := strconv.Atoi(pos[0:2]); err == nil {
			qnnumber = i2
			rest = pos[2:]
		} else {
			if i1, err := strconv.Atoi(pos[0:1]); err == nil {
				qnnumber = i1
				rest = pos[1:]
			}
		}
	} else {
		if i1, err := strconv.Atoi(pos[0:1]); err == nil {
			qnnumber = i1
			rest = pos[1:]
		}
	}

	return
}

// lastPos must either be "", then pos must be complete
// (i.e. must start with a number) or lastPos must be complete
// then pos may be derived from it
func positionTo32th(lastPos, pos string) (completed string, num32th uint) {

	number, rest := getQNNumberFromPos(pos)
	completed = pos

	if number == -1 {
		if lastPos == "" {
			panic("lastPos must be given, if pos is incomplete")
		}

		lastNum, lastRest := getQNNumberFromPos(lastPos)

		if lastNum < 1 {
			panic("lastPos must be given, if pos is incomplete")
		}

		number = lastNum
		rest = lastRest + rest
		completed = fmt.Sprintf("%v%s", number, rest)
	}

	num32th = uint((number - 1) * 8)

	if rest == "" {
		return
	}

	switch rest {
	case ";":
		num32th += 1
	case ".":
		num32th += 2
	case ".;":
		num32th += 3
	case "&":
		num32th += 4
	case "&;":
		num32th += 5
	case "&.":
		num32th += 6
	case "&.;":
		num32th += 7
	default:
		panic("invalid rest: " + rest + " in " + pos)
	}

	return

}
