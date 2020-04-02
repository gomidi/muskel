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

func _applyLyric(ev *Event, lyric string) (applied *Event, wasApplied bool) {
	applied = ev.Dup()
	var l Lyric
	l.Text = lyric

	switch v := applied.Item.(type) {
	case *MultiItem:
		if !MultiHasNote(v) {
			return
		}
		var _ev = ev.Dup()
		_ev.PosShift = 0
		_ev.Item = &l
		v.Events = append(v.Events, _ev)
		applied.Item = v
		wasApplied = true
	case *Note:
		var me = &MultiItem{}
		me.Reset()
		//l.PosShift = v.PosShift
		_ev := ev.Dup()
		_ev.PosShift = 0
		_ev.Item = &l

		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	case *MIDINote:
		var me = &MultiItem{}
		me.Reset()
		_ev := ev.Dup()
		_ev.PosShift = 0
		_ev.Item = &l
		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	case *NTuple:
		if !NtupleHasNote(v) {
			return
		}
		var me = &MultiItem{}
		me.Reset()
		var _ev = ev.Dup()
		_ev.Item = &l
		_ev.PosShift = 0
		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		//l.PosShift = v.PosShift
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	default:
		if ev.Item != Hold {
			return
		}

		var me = &MultiItem{}
		me.Reset()
		_ev := ev.Dup()
		_ev.Item = &l
		_ev.PosShift = 0
		_ev2 := ev.Dup()
		_ev2.PosShift = 0
		me.Events = append(me.Events, _ev2, _ev)
		applied.Item = me
		wasApplied = true
	}
	return
}

// applyLyrics applies the lyrics to the items of the pattern
func ApplyLyrics(events []*Event, lyrics []string) (applied []*Event) {
	if len(lyrics) == 0 {
		return events
	}
	var i int
	for _, ev := range events {
		if i >= len(lyrics) {
			i = i % len(lyrics)
		}
		nev, wasApplied := _applyLyric(ev, lyrics[i])
		if wasApplied {
			i++
		}
		applied = append(applied, nev)
	}
	return
}

func MultiHasNote(me *MultiItem) bool {
	for _, it := range me.Events {
		switch it.Item.(type) {
		case *Note, *MIDINote:
			return true
		default:
		}
	}
	return false
}

func NtupleHasNote(me *NTuple) bool {
	for _, it := range me.Events {
		switch it.Item.(type) {
		case *Note, *MIDINote:
			return true
		default:
		}
	}
	return false
}

func IsLoop(it Item) int {
	if it == nil {
		return -1
	}

	switch v := it.(type) {
	case *Override:
		return IsLoop(v.Item)
	case *BarRepeater:
		if v.OnlyOnce {
			return 0
		}
		return v.LastN
	default:
		return 0
	}
}

func FindEventThatIsNotRest(start int, evts []*Event) (found int) {
	if start >= len(evts) {
		return -1
	}

	if evts[start].Item != Rest {
		return start
	}

	return FindEventThatIsNotRest(start+1, evts)
}

//func replaceParams(s string, params []interface{}) string {
func ReplaceParams(s string, params []string) string {
	/*
		if DEBUG {
			fmt.Printf("should replace params %#v in %q\n", params, s)
		}
	*/
	if len(params) == 0 {
		return s
	}
	paramslen := len(params)
	return TemplateReg.ReplaceAllStringFunc(s, func(in string) (out string) {
		i, err := strconv.Atoi(in[1:])
		if err != nil {
			// should be unreachable, if regexp is correct
			panic(in)
		}
		return params[(i-1)%paramslen]
	})
}

// sliceEvents slices the events according to the template call slice definition
func SliceEvents(slice [2]int, all []*Event, projectedBarEnd uint) (evs []*Event, absoluteEnd uint) {
	first, last := slice[0], slice[1]
	absoluteEnd = projectedBarEnd
	if first < 0 {
		return all, projectedBarEnd
	}

	var collecting bool

	for i, ev := range all {
		if last > 0 && i >= last {
			absoluteEnd = ev.Position
			break
		}

		if i == first {
			collecting = true
		}

		if collecting {
			evs = append(evs, ev.Dup())
		}
	}
	return
}

func reduceDynamic(dyn string) string {

	var counter int
	for _, sym := range dyn {
		switch sym {
		case '-':
			counter--
		case '+':
			counter++
		}
	}

	//fmt.Printf("reduceDynamic: %q: %v\n", dyn, counter)

	switch {
	case counter > 0:
		return strings.Repeat("+", counter)
	case counter < 0:
		return strings.Repeat("-", counter*(-1))
	default:
		return ""
	}
}

func AddDynamic(orig, dynAdd string) (nu string) {
	//fmt.Printf("addDynamic %q to %q\n", p.DynamicAdd, orig)
	if orig == "=" {
		return orig
	}

	if dynAdd == "" {
		return orig
	}

	if dynAdd == "=" {
		return dynAdd
	}

	return reduceDynamic(orig + dynAdd)
}

func calcAdd(distance int64, diff float64) float64 {
	return diff / float64(distance)
}

func calcAdd2(distance int64, diff float64) float64 {
	return diff / (float64(distance) * float64(distance))
}

var DottedLengths = map[string][2]uint32{
	//".":   [2]uint32{1, 8},
	":":   [2]uint32{1, 16},
	"::":  [2]uint32{1, 32},
	":::": [2]uint32{1, 64},
}

func VelocityToDynamic(vel uint8) (dyn string) {
	switch {
	case vel >= 119:
		return "++++"
	case vel >= 104:
		return "+++"
	case vel >= 89:
		return "++"
	case vel >= 74:
		return "+"
	case vel >= 59:
		return "="
	case vel >= 44:
		return "-"
	case vel >= 29:
		return "--"
	case vel >= 14:
		return "---"
	default:
		return "----"
	}
}

func LinearGlide(distance int64, noteDiff int64, callback func(step uint, val float64)) {

	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance * step
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance * distance = note0 + noteTarget - note0 = noteTarget
	// y             f(x)        =  a     +      m                      * x
	// m                         =  (noteTarget-note0)/distance

	m := calcAdd(distance, float64(noteDiff))

	for step := int64(1); step <= distance; step++ {
		callback(uint(step), m*float64(step))
	}
}

func CalcNoteDelay(resolution smf.MetricTicks) (delay int32) {
	return int32(math.Round(float64(resolution.Ticks4th()) * 4 / 128))
}

func ExponentialGlide(distance int64, noteDiff int64, callback func(pos uint, val float64)) {
	// y             f(x)        =  a     +      m                       * x²
	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance² * step²
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance² * distance² = note0 + noteTarget - note0 = noteTarget
	// m                         =  (noteTarget-note0)/distance²

	m := calcAdd2(distance, float64(noteDiff))

	for step := int64(1); step <= distance; step++ {
		callback(uint(step), m*float64(step)*float64(step))
	}
}

func LinearTempoChange(distance int64, diff float64, callback func(step uint, val float64)) {

	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance * step
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance * distance = note0 + noteTarget - note0 = noteTarget
	// y             f(x)        =  a     +      m                      * x
	// m                         =  (noteTarget-note0)/distance

	m := calcAdd(distance, diff)

	for step := int64(1); step <= distance; step++ {
		callback(uint(step), m*float64(step))
	}
}

func ExponentialTempoChange(distance int64, diff float64, callback func(step uint, val float64)) {
	// y             f(x)        =  a     +      m                       * x²
	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance² * step²
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance² * distance² = note0 + noteTarget - note0 = noteTarget
	// m                         =  (noteTarget-note0)/distance²

	m := calcAdd2(distance, diff)
	for step := int64(1); step <= distance; step++ {
		callback(uint(step), m*float64(step)*float64(step))
	}
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

// syntax for params:  #1 #2 etc.
var regPos = regexp.MustCompile("^([1-9]?[0-9]?)([&;" + regexp.QuoteMeta(",") + regexp.QuoteMeta(",") + "]*)")

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
		fmt.Fprint(&bf, ",")
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

func GetQNNumberFromPos(pos string) (qnnumber uint, rest string) {
	if len(pos) == 0 {
		panic("empty position is not valid")
	}
	qnnumber = 0
	rest = pos

	if len(pos) > 1 {
		if i2, err := strconv.Atoi(pos[0:2]); err == nil {
			qnnumber = uint(i2)
			rest = pos[2:]
		} else {
			if i1, err := strconv.Atoi(pos[0:1]); err == nil {
				qnnumber = uint(i1)
				rest = pos[1:]
			}
		}
	} else {
		if i1, err := strconv.Atoi(pos[0:1]); err == nil {
			qnnumber = uint(i1)
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

/*
lastBeatNo is the number of the beat in the bar, starting with 1
if lastBeatNo = 0, then there is no lastBeatNo given
*/
func PositionTo32th(lastBeatNo uint, pos string) (completed string, num32th uint, err error) {

	number, rest := GetQNNumberFromPos(pos)
	completed = pos

	if number == 0 {
		if lastBeatNo < 0 {
			err = fmt.Errorf("lastBeatNo must be given, if pos is incomplete")
			return
		}

		number = lastBeatNo
		completed = fmt.Sprintf("%v%s", number, rest)

		//		fmt.Printf("lastPos: %q pos: %q completed: %q\n", lastPos, pos, completed)

	}

	num32th = uint((number - 1) * 8)

	if rest == "" {
		return
	}

	if rest[0] == '.' {
		var fl float64
		fl, err = strconv.ParseFloat("0"+rest, 64)

		if err != nil {
			err = fmt.Errorf("invalid float position %q", pos)
			return
		}

		switch fl {
		case 0.125:
			num32th += 1
		case 0.25:
			num32th += 2
		case 0.375:
			num32th += 3
		case 0.5:
			num32th += 4
		case 0.625:
			num32th += 5
		case 0.75:
			num32th += 6
		case 0.875:
			num32th += 7
		default:
			err = fmt.Errorf("invalid float position %q", pos)
		}
		return
	}

	switch rest {
	case ";":
		num32th += 1
	case ",":
		num32th += 2
	case ",;":
		num32th += 3
	case "&":
		num32th += 4
	case "&;":
		num32th += 5
	case "&,":
		num32th += 6
	case "&,;":
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
