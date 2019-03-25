package muskel

import (
	"fmt"
	"os"
	"path/filepath"
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
	var bf strings.Builder

	bf.WriteString(s)

	for i := 0; i < diff; i++ {
		bf.WriteString(" ")
	}
	return bf.String()
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func findInclude(file string) (resolved string, err error) {
	if fileExists(file) {
		return filepath.Abs(file)
	}

	if filepath.IsAbs(file) {
		return "", fmt.Errorf("file not found: %q", file)
	}

	try1 := filepath.Join(WORKING_DIR, file)
	if fileExists(try1) {
		return filepath.Abs(try1)
	}

	try2 := filepath.Join(USER_DIR, file)
	if fileExists(try2) {
		return filepath.Abs(try2)
	}

	return "", fmt.Errorf("file not found: %q", file)
}

// if returned vel is < 0, then it has not been set and
// is drived from the previous velocity
// always return such that rand.Intn(4) can be added or substracted
func dynamicToVelocity(dyn string) (vel int8) {
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
		case '=': // reset
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

func velocityToDynamic(vel int8) (dyn string) {
	switch vel {
	case -1:
		return ""
	case 63:
		return "="
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
	if regPos.MatchString(s) {
		all := regPos.FindAllString(s, 1)
		f.position = all[0]

		s = regPos.ReplaceAllString(s, "")
	}
	f.item = s
}

// syntax for params:  #1 #2 etc.
var templateReg = regexp.MustCompile(regexp.QuoteMeta("#") + "([1-9]+)")
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
	var frepl templateFragment
	frepl.parse(replacement)

	position = frepl.position

	item = frepl.item

	return
}

func pos32thToString(pos uint) string {
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
func positionTo32th(lastPos, pos string) (completed string, num32th uint, err error) {

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
