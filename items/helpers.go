package items

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/smf"
)

func calcAdd(distance int64, diff float64) float64 {
	return diff / float64(distance)
}

func calcAdd2(distance int64, diff float64) float64 {
	return diff / (float64(distance) * float64(distance))
}

var dottedLengths = map[string][2]uint32{
	//".":   [2]uint32{1, 8},
	":":   [2]uint32{1, 16},
	"::":  [2]uint32{1, 32},
	":::": [2]uint32{1, 64},
}

func linearGlide(wr SMFWriter, distance int64, noteDiff int64, callback func(val float64)) {

	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance * step
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance * distance = note0 + noteTarget - note0 = noteTarget
	// y             f(x)        =  a     +      m                      * x
	// m                         =  (noteTarget-note0)/distance

	m := calcAdd(distance, float64(noteDiff))
	wr.RestoreTimeline()

	for step := int64(1); step <= distance; step++ {
		MIDITrack.GlideForward(wr)
		callback(m * float64(step))
	}
}

func CalcNoteDelay(resolution smf.MetricTicks) (delay int32) {
	return int32(math.Round(float64(resolution.Ticks4th()) * 4 / 128))
}

func exponentialGlide(wr SMFWriter, distance int64, noteDiff int64, callback func(val float64)) {
	// y             f(x)        =  a     +      m                       * x²
	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance² * step²
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance² * distance² = note0 + noteTarget - note0 = noteTarget
	// m                         =  (noteTarget-note0)/distance²

	m := calcAdd2(distance, float64(noteDiff))
	wr.RestoreTimeline()

	for step := int64(1); step <= distance; step++ {
		MIDITrack.GlideForward(wr)
		callback(m * float64(step) * float64(step))
	}
}

func LinearTempoChange(wr SMFWriter, distance int64, diff float64, callback func(val float64)) {

	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance * step
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance * distance = note0 + noteTarget - note0 = noteTarget
	// y             f(x)        =  a     +      m                      * x
	// m                         =  (noteTarget-note0)/distance

	m := calcAdd(distance, diff)
	//fmt.Printf("linearGlissando: m = %0.5f\n", m)
	wr.RestoreTimeline()
	//var pb int16

	for step := int64(1); step <= distance; step++ {
		//iw.wr.Forward(0, 1, 32)
		wr.Forward(0, 1, 16)
		callback(m * float64(step))
		//pb = halfTonesToPitchbend(m*float64(step), iw.instr.PitchbendRange)
		//iw.wr.Pitchbend(pb)
	}

	//iw.prevPitchbend = pb
}

func ExponentialTempoChange(wr SMFWriter, distance int64, diff float64, callback func(val float64)) {
	// y             f(x)        =  a     +      m                       * x²
	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance² * step²
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance² * distance² = note0 + noteTarget - note0 = noteTarget
	// m                         =  (noteTarget-note0)/distance²

	m := calcAdd2(distance, diff)
	wr.RestoreTimeline()
	//var pb int16

	for step := int64(1); step <= distance; step++ {
		//iw.wr.Forward(0, 1, 32)
		wr.Forward(0, 1, 16)
		callback(m * float64(step) * float64(step))
		//pb = halfTonesToPitchbend(m*float64(step)*float64(step), iw.instr.PitchbendRange)
		//iw.wr.Pitchbend(pb)
	}

	//iw.prevPitchbend = pb
}

func stripNoteOnOff(data string) (noteOn bool, noteOff bool, rest string) {
	rest = data

	if usIdx := strings.LastIndex(rest, "_"); usIdx == len(rest)-1 && usIdx > 0 {
		noteOn = true
		rest = rest[:usIdx]
	}

	if usIdx := strings.Index(rest, "_"); usIdx == 0 {
		noteOff = true
		rest = rest[1:]
	}

	return
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

// if returned vel is < 0, then it has not been set and
// is drived from the previous velocity
// always return such that rand.Intn(4) can be added or substracted
func DynamicToVelocity(dyn string, min, max, randomness, step, center int8) (vel int8) {

	//vel = -1
	// reset to mezzoforte
	//vel = 63
	vel = center

	for _, rn := range dyn {
		switch rn {
		case '+':
			/*
				if vel == -1 {
					vel = 63
				}
			*/
			vel += int8(step)
		case '-':
			/*
				if vel == -1 {
					vel = 63
				}
			*/
			vel -= int8(step)
		case '=':
			//vel = 63
			// take previous value
			vel = -1
		}
	}

	if vel == -1 {
		return vel
	}

	if vel > max-randomness {
		vel = max - randomness
	}

	if vel < min+randomness {
		vel = min + randomness
	}

	return
}

func CentsToPitchbend(cents float64, _range uint8) int16 {
	return HalfTonesToPitchbend(cents/100.0, _range)
}

func HalfTonesToPitchbend(halftones float64, _range uint8) int16 {
	res := int16(math.Round(halftones * 8192.0 / float64(_range)))
	if res < -8191 {
		return -8191
	}

	if res > 8192 {
		return 8192
	}

	return res
}

func FreqRatioToCent(num, denom uint8) float64 {
	// 1200·log2(6⁄5) Cent = 315,641 Cent

	//log2(6/5) = ln(6/5) / ln(2)

	return 1200.0 * math.Log2(float64(num)/float64(denom))
}

func velocityToDynamic(vel int8) (dyn string) {
	switch vel {
	case -1:
		//return ""
		return "="
	case 63:
		//return "="
		return ""
	case 78:
		return "+"
	case 93:
		return "++"
	case 108:
		return "+++"
	case 123:
		return "++++"
	case 48:
		return "-"
	case 33:
		return "--"
	case 18:
		return "---"
	case 5:
		return "----"
	default:
		panic(fmt.Sprintf("invalid value for velocity: %v", vel))
	}
}

type templateFragment struct {
	position string
	item     string
}

func (f *templateFragment) parse(s string) {
	//fmt.Printf("templateFragment.parse(%q)\n", s)
	if regPos.MatchString(s) {
		all := regPos.FindAllString(s, 1)
		//fmt.Printf("matching: %v\n", all)
		f.position = all[0]

		s = regPos.ReplaceAllString(s, "")
	}
	f.item = s
}

// syntax for params:  #1 #2 etc.
var regPos = regexp.MustCompile("^([1-9]?[0-9]?)([&;" + regexp.QuoteMeta(".") + "]*)")

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
	var frepl templateFragment
	frepl.parse(replacement)

	position = frepl.position

	item = frepl.item

	return
}

func Pos32thToString(pos uint) string {
	qn := pos / 8
	rest := pos % 8

	var bf strings.Builder
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
func Length32ths(num, denom uint8) int {
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

func splitParams(params string) (res []string) {
	var embed int
	var bf bytes.Buffer

	for _, char := range params {
		switch char {
		case ',':
			if embed > 0 {
				bf.WriteRune(char)
			} else {
				res = append(res, bf.String())
				bf.Reset()
			}
		case '{':
			embed++
			bf.WriteRune(char)
		case '}':
			embed--
			bf.WriteRune(char)
		default:
			bf.WriteRune(char)
		}
	}
	res = append(res, bf.String())
	return
}

// lastPos must either be "", then pos must be complete
// (i.e. must start with a number) or lastPos must be complete
// then pos may be derived from it
func PositionTo32th(lastPos, pos string) (completed string, num32th uint, err error) {

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

		//		fmt.Printf("lastPos: %q pos: %q completed: %q\n", lastPos, pos, completed)

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
		err = fmt.Errorf("invalid rest: %q in position %q", rest, pos)
	}

	return

}

func slice(start, end int, s [][]bool) [][]bool {
	if end < 0 {
		end = len(s) + end
	}
	return append([][]bool{}, s[start:end]...)
}

func mapp(gr [][]bool, fn func([]bool, int) []bool) [][]bool {
	var result [][]bool
	for i, gg := range gr {
		result = append(result, fn(gg, i))
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compareArrays(a, b []bool) (equal bool) {
	return reflect.DeepEqual(a, b)
}
