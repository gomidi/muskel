package items

import (
	"fmt"
	"strings"
)

func runeItem(r rune) Item {
	switch r {
	case '%':
		return repeatLastEvent{}
	case ':':
		return Hold
	case '*':
		return Rest
	case '~':
		return &GlideLinear
	case '_':
		return End
	default:
		return &Note{}
	}
}

func getMIDIItem(data string) Item {
	switch data[0:2] {
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
	default:
		return nil
	}
}

func getItem(data string) Item {
	switch data[0] {
	case '>':
		return &MidiSample{}
	case '\'':
		return &CellReference{}
	case '[':
		// EventSequence
		return &PartRepeat{}
	case '#':
		var p Placeholder
		return &p
	case '~':
		return &GlideLinear
	case '"':
		return &Lyric{}
	case '{':
		return &NTuple{}
	case '$':
		piidx := strings.Index(data, "/")
		if data[1] != '$' && piidx > 0 {
			return &PipedPatternCommands{}
		}
		return &Command{}
	case '(':
		// ItemGroup
		return &MultiItem{}
	case '\\':
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
		switch {
		case len(data) == 3 && data[2] == '.':
			// EventSequence
			return &BarRepeater{}
		case regExToken0.MatchString(data) || regExToken1.MatchString(data):
			// ItemGroup
			return &Token{}
		default:
			return nil
		}
	}

	switch data[0:2] {
	//case "_°": // hack for keyboads that don't allow to properly insert ^
	case "MN", "CC", "PB", "PC", "AT", "PT":
		return getMIDIItem(data)
	}

	switch {
	case len(data) > 3 && data[0:3] == "_MN":
		return &MIDINote{}
	case regExEndScaleNote.MatchString(data):
		return &Note{}
	case regExTemplate0.MatchString(data) || regExTemplate1.MatchString(data):
		// EventSequence
		return &Pattern{}
	default:
		return &Note{}
	}
}

// Parse is the major entry point for parsing an Item
func Parse(data string, posIn32th uint) (it Item, err error) {
	data = strings.TrimSpace(data)

	if DEBUG {
		fmt.Printf("Parse(%q)\n", data)
	}

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

	if cmd, is := it.(*Command); is && cmd.IsInclude() {
		return cmd.GetInclude(), nil
	}

	return it, nil
}
