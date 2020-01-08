package items

import (
	"fmt"
	"strconv"
	"strings"
)

type TuningCents struct {
	Name     string
	Original string
	notes    []uint8
	cents    []float64
}

func (m *TuningCents) findIndex(what uint8) (idx int) {
	//fmt.Printf("looking for %v\n", what)

	idx = -1

	for _idx, nt := range m.notes {
		if nt == what {
			return _idx
		}
	}

	return idx
}

func (m *TuningCents) findTuning(note uint8) int {
	for i := note % 12; i < 127; i += 12 {
		idx := m.findIndex(i)

		if idx > -1 {
			return idx
		}
	}

	return -1
}

func (m *TuningCents) tryTunings(midinote uint8) int {
	idx := m.findIndex(midinote)

	if idx > -1 {
		return idx
	}

	if idx := m.findTuning(midinote); idx > -1 {
		return idx
	}

	if idx := m.findTuning(midinote + 1); idx > -1 {
		return idx
	}

	if midinote > 0 {
		if idx := m.findTuning(midinote - 1); idx > -1 {
			return idx
		}
	}

	if idx := m.findTuning(midinote + 2); idx > -1 {
		return idx
	}

	if midinote > 1 {
		if idx := m.findTuning(midinote - 2); idx > -1 {
			return idx
		}
	}

	if idx := m.findTuning(midinote + 3); idx > -1 {
		return idx
	}

	if midinote > 2 {
		if idx := m.findTuning(midinote - 3); idx > -1 {
			return idx
		}
	}

	panic("can't find tuning within +-3 halftone range")
}

func (m *TuningCents) CentsFor(midinote uint8) float64 {
	idx := m.tryTunings(midinote)

	if idx > -1 {
		return m.cents[idx]
	}

	panic("can't find tuning within +-3 halftone range")
}

func (t *TuningCents) Pitchbend(midinote uint8, _range uint8) int16 {
	return CentsToPitchbend(t.CentsFor(midinote), _range)
}

type Tuning interface {
	Pitchbend(midinote uint8, _range uint8) int16
}

type TuningRatio struct {
	baseFreq float64
	Name     string
	Original string
	notes    []uint8
	ratios   [][2]uint8
	center   uint8
}

func (m *TuningRatio) findIndex(what uint8) (idx int) {
	//fmt.Printf("looking for %v\n", what)

	idx = -1

	for _idx, nt := range m.notes {
		if nt == what {
			return _idx
		}
	}

	return idx
}

func (m *TuningRatio) findTuning(note uint8) int {
	for i := note % 12; i < 127; i += 12 {
		idx := m.findIndex(i)

		if idx > -1 {
			return idx
		}
	}

	return -1
}

func (m *TuningRatio) tryTunings(midinote uint8) int {
	idx := m.findIndex(midinote)

	if idx > -1 {
		return idx
	}

	if idx := m.findTuning(midinote); idx > -1 {
		return idx
	}

	if idx := m.findTuning(midinote + 1); idx > -1 {
		return idx
	}

	if midinote > 0 {
		if idx := m.findTuning(midinote - 1); idx > -1 {
			return idx
		}
	}

	if idx := m.findTuning(midinote + 2); idx > -1 {
		return idx
	}

	if midinote > 1 {
		if idx := m.findTuning(midinote - 2); idx > -1 {
			return idx
		}
	}

	if idx := m.findTuning(midinote + 3); idx > -1 {
		return idx
	}

	if midinote > 2 {
		if idx := m.findTuning(midinote - 3); idx > -1 {
			return idx
		}
	}

	panic("can't find tuning within +-3 halftone range")
}

func (m *TuningRatio) RatioFor(midinote uint8) [2]uint8 {
	idx := m.tryTunings(midinote)

	if idx > -1 {
		return m.ratios[idx]
	}

	panic("can't find tuning within +-3 halftone range")
}

func (t *TuningRatio) Parse(name, baseFreq, def string) (err error) {
	t.Name = name
	t.Original = def

	t.baseFreq, err = strconv.ParseFloat(baseFreq, 64)

	if err != nil {
		return fmt.Errorf("baseFreq of tuning %q is not a float", name)
	}

	defs := strings.Split(def, ",")

	for idx, pair := range defs {
		pr := strings.Split(pair, "=")

		if len(pr) != 2 {
			return fmt.Errorf("error in tuning definition %q at position %v: %q is no valid tuning pair", t.Name, idx, pair)
		}

		rat := strings.Split(pr[1], "/")

		num := rat[0]
		denom := "1"
		if len(rat) > 1 {
			denom = rat[1]
		}

		nn, err := strconv.Atoi(num)

		if err != nil {
			return fmt.Errorf("error in tuning definition %q at position %v: %q is no valid ratio", t.Name, idx, pr[1])
		}

		dn, err := strconv.Atoi(denom)

		if err != nil {
			return fmt.Errorf("error in tuning definition %q at position %v: %q is no valid ratio", t.Name, idx, pr[1])
		}

		nt, err := ParseNote(pr[0])

		if err != nil {
			return fmt.Errorf("error in tuning definition %q at position %v: %q is no valid note", t.Name, idx, pr[0])
		}

		//t.tuning = append(t.tuning, [2]int{int(nt.ToMIDI()), cents})
		t.notes = append(t.notes, nt.ToMIDI())
		t.ratios = append(t.ratios, [2]uint8{uint8(nn), uint8(dn)})

		if nn == 1 && dn == 1 {
			t.center = nt.ToMIDI()
		}
	}

	return nil
}

func (t *TuningRatio) Pitchbend(midinote uint8, _range uint8) int16 {
	rat := t.RatioFor(midinote)
	cents := FreqRatioToCent(rat[0], rat[1])

	dist := int(midinote) - int(t.center)
	cents -= float64(dist * 100)

	return CentsToPitchbend(cents, _range)
}

func ParseTuning(data string) (Tuning, error) {
	idx := strings.Index(data, ":")

	if idx < 1 || idx >= len(data)-1 {
		return nil, fmt.Errorf("error in tuning syntax: : missing")
	}

	name := strings.TrimSpace(data[:idx])
	def := strings.TrimSpace(data[idx+1:])

	idxRat := strings.Index(name, "@")

	if idxRat < 1 {
		var tn TuningCents
		err := tn.Parse(name, def)

		if err != nil {
			return nil, err
		}

		return &tn, nil
	}

	var tr TuningRatio
	err := tr.Parse(name[:idxRat], name[idxRat+1:], def)

	if err != nil {
		return nil, err
	}

	return &tr, nil

}

func (t *TuningCents) Parse(name, def string) error {
	t.Name = name
	t.Original = def

	defs := strings.Split(def, ",")

	for idx, pair := range defs {
		pr := strings.Split(pair, ":")

		if len(pr) != 2 {
			return fmt.Errorf("error in tuning definition %q at position %v: %q is no valid tuning pair", t.Name, idx, pair)
		}

		cents, err := strconv.ParseFloat(pr[1], 64)

		if err != nil {
			return fmt.Errorf("error in tuning definition %q at position %v: %v is no valid cent number", t.Name, idx, pr[1])
		}

		if cents > 100 || cents < -100 {
			return fmt.Errorf("error in tuning definition %q at position %v: %v is no valid cent number", t.Name, idx, pr[1])
		}

		nt, err := ParseNote(pr[0])

		if err != nil {
			return fmt.Errorf("error in tuning definition %q at position %v: %q is no valid note", t.Name, idx, pr[0])
		}

		//t.tuning = append(t.tuning, [2]int{int(nt.ToMIDI()), cents})
		t.notes = append(t.notes, nt.ToMIDI())
		t.cents = append(t.cents, cents)
	}

	return nil
}

/*
Für die Definition eines Tunings können wir nun ähnlich vorgehen.

Hier fangen wir mit der Raute vor dem Namen des Tunings an und haben eine Zuordnung von absoluten Noten zu entweder cent-Differenzen
zur entsprechenden Frequenzen der absoluten Note, oder direkt in absoluten Frequenzwerten. Die cent-Werte werden hinter dem Doppelpunkt angegeben, Frequenzen hinter dem =:

#my_tune:  d":23,d#":12,e":-12,f#":40,g":30,a"=400,c":10

Angegebene Frequenzen werden direkt in cent differenzen() umgerechnet und intern werden nur die cent Differenzen gespeichert.
Die Auswahl der entsprechenden Cent-Differenz einer absoluten Note erfolgt nach dem gleichen Prinzip des "nächsten Tones" und der
Oktavierungen (oder eben keiner), wie bei den Skalendefinitionen. Hier wird als Cent-Differenz eines absoluten Tones immer
die Cent-Differenz des "nächsten Tones" (sei er auch oktaviert) verwendet, wobei als Nähe die Anzahl der Halbtonschritte gemäß
Benennung genommen wird. Wenn man alle Halbtöne in einem Tuning definiert, hat man also die größte Kontrolle.


Es wäre noch zu Überlegen,ob es für Tuning und Skalendefinition nicht nett wäre, wenn die Viertel-Tönige und Achteltönige
Notation für die absolten Noten Verwendung finden würde. Damit mehrere kleinabstände innerhalb einer Skala möglich wären:

c    = ganzton c
c^   = achtelton zwischen c und c°
c°   = viertelton zwischen c und c#
c°^  = achtelton zwischen c° und c#
c#   = halbton c#
c#^  = achtelton zwischen c# und c#°
c#°  = viertelton zwischen c# und d
c#°^ = achtelton zwischen c#° und d
d    = ganzton d


*/
