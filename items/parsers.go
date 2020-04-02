package items

import (
	"fmt"
	"strings"
)

func runeItem(r rune) Item {
	switch r {
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		//err = nt.parseScale(data)
		return &Note{}
	case '%':
		return repeatLastEvent{}
	case ':':
		return Hold
	case '*':
		return Rest
	case '~':
		gliss := GlideStart(false)
		return &gliss
	case '_':
		return End
	default:
		return &Note{}
	}
}

func getItem(data string) (it Item) {
	switch data[0] {
	case '[':
		// EventSequence
		return &PartRepeat{}
	case '#':
		var p Placeholder
		return &p
	case '~':
		gliss := GlideStart(false)
		return &gliss
	case '"':
		return &Lyric{}
	case '{':
		return &NTuple{}
	case '$':
		piidx := strings.Index(data, "/")
		if data[1] != '$' && piidx > 0 {
			return &PipedPatternCommands{}
		}
		//fmt.Printf("got command %q\n", data)
		return &Command{}
	case '(':
		// ItemGroup
		return &MultiItem{}
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return &Note{}
	case '\\': // scale
		// ItemGroup
		return &Scale{}
	case '?':
		if data[1] == '(' {
			return &RandomChooser{}
		}
		return &RandomProbability{}
	case '/':
		return &Override{}
	case '.':
		if len(data) == 3 && data[2] == '.' {
			// EventSequence
			return &BarRepeater{}
		}

		if regExToken0.MatchString(data) || regExToken1.MatchString(data) {
			// ItemGroup
			return &Token{}
		}

		return nil

	default:

		if len(data) > 3 && data[0:3] == "_MN" {
			return &MIDINote{}
		}

		switch data[0:2] {
		//case "_°": // hack for keyboads that don't allow to properly insert ^
		case "MN":
			return &MIDINote{}
		case "CC":
			return &MIDICC{}
		case "PB":
			return &MIDIPitchbend{}
		case "PC":
			return &MIDIProgramChange{}
		case "AT":
			return &MIDIAftertouch{}
		case "PT":
			return &MIDIPolyAftertouch{}
		}

		if regExEndScaleNote.MatchString(data) {
			//err = nt.parseScale(data)
			return &Note{}
		}

		if regExTemplate0.MatchString(data) || regExTemplate1.MatchString(data) {
			// EventSequence
			return &Pattern{}
		}

		return &Note{}
	}
}

func Parse(data string, posIn32th uint) (it Item, err error) {
	data = strings.TrimSpace(data)

	switch len(data) {
	case 0:
		return nil, nil
	case 1:
		it = runeItem(rune(data[0]))
	default:
		it = getItem(data)
	}

	if it == nil {
		return nil, fmt.Errorf("unknown item %T: %q", it, data)
	}

	err = it.Parse(data, posIn32th)
	if err != nil {
		return nil, fmt.Errorf("invalid item %T: %q", it, data)
	}

	/*
		if cmd, is := it.(*Command); is {
			fmt.Printf("cmd: %#v\n", cmd)
		}
	*/

	if cmd, is := it.(*Command); is && cmd.Name == "$include" {
		inc := Include{
			File:   strings.Trim(cmd.Params[0], `"'`),
			Sketch: "=SCORE",
		}
		if len(cmd.Params) > 1 {
			inc.Sketch = strings.Trim(cmd.Params[1], `"'`)
		}

		if len(cmd.Params) > 2 {
			inc.Params = cmd.Params[2:]
		}
		return inc, nil
	}

	return it, nil
}
