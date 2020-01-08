package items

import (
	"fmt"
	"strings"
)

func abs(x int) int {
	if x < 0 {
		return x * (-1)
	}
	return x
}

func (m *ScaleDefinition) findIndex(what int) (idx int) {
	idx = -1

	for _idx, nt := range m.notes {
		mn := int(nt.ToMIDI())

		if mn == what {
			return _idx
		}
	}

	return idx
}

func (m *ScaleDefinition) findStep(basenote uint8, note int) (step int8) {
	for i := note % m.span; i < 127; i += m.span {
		idx := m.findIndex(i)

		if idx > -1 {
			//fmt.Printf("found: %v at %v\n", i, idx)
			//m.zeroIdx = idx
			//m.halftones = 12 * (i%12 - fac)
			diff := idx - m.zeroIdx
			oct := (i - int(basenote)) / m.span

			if diff > 0 {
				return int8(diff + 1 + oct)
			} else {
				oct += 1
				return int8((m.span*oct)+diff-m.zeroIdx) + 1
			}
		}
	}

	// 0 is not allowed step == not existing
	return 0
}

/*
\my_scale:                                       d",e",f#",g",a",c"

resultiert in (halbton-differenz zum Nullpunkt)  -7,-5,-3 ,-2, 0,3
*/

/*
type Mode interface {
	Name() string
	NoteToStep(basenote, note uint8) (step int8)
	StepToNote(basenote uint8, step int8) (note uint8)
}

*/

var _ Mode = &ScaleDefinition{}

type ScaleDefinition struct {
	name       string
	definition []int
	Original   string
	notes      []Note
	zeroIdx    int
	numSteps   int
	span       int
	center     int
}

func (s *ScaleDefinition) All(basenote uint8) (notes []uint8) {
	for i := 0; i < s.numSteps; i++ {
		notes = append(notes, s.StepToNote(basenote, int8(i)))
	}
	return
}

func (s *ScaleDefinition) Name() string {
	return s.name
}

func (s *ScaleDefinition) NoteToStep(basenote, note uint8) (step int8) {
	switch {
	case note == basenote:
		return 1
	default:
		return s.findStep(basenote, int(note))
	}
}

func keyToNote(key uint8) (nt Note) {
	nt.Letter, nt.Augmenter, nt.Octave = KeyToNote(key)
	return
}

func (s *ScaleDefinition) StepToNote(basenote uint8, step int8) (note uint8) {
	var x int
	if step >= 0 {
		step += 1
	}

	//fmt.Printf("StepToNote(basenote: %s, step %v)\n", keyToNote(basenote), step)

	switch {
	case step == 1:
		return basenote
	case step > 1:
		x = int(step) - 1 + s.zeroIdx
	case step < 0:
		x = s.zeroIdx + int(step)
	default:
		panic("step 0 not allowed")
	}

	reduced, fact := x%s.numSteps, x/s.numSteps
	if reduced < 0 {
		reduced += s.numSteps
		fact -= 1
	}

	//fmt.Printf("StepToNote reduced: %v fact: %v span: %v\n", reduced, fact, s.span)

	return uint8(int(basenote) + s.definition[reduced] + fact*s.span)
}

/*
\my_scale:                                       d",e",f#",g",a",c"

resultiert in (halbton-differenz zum Nullpunkt)  -7,-5,-3 ,-2, 0,3


Die absoluten Noten werden aber intern auch gespeichert.
Beim verwenden einer Skala, wird der neue Nullpunkt anhand einer absoluten Note gesetzt. Hierbei wird zunächst
die "nächste" absolute Note gemäß der Definition gesucht, gemerkt und die entsprechende halbtondifferenz gemerkt.
Bei der Suche der nächsten Note wird wie folgt vorgegangen:
Der eingehängte Ton wird solange oktaviert, bis die Differenz zu einer der bei der Definition der Skala
gesetzten absoluten Note minimal ist. Angestrebt werden 0-3 Halbtöne. Dann wird dieser "nächste" Ton und die
Differenz dazu gemerkt.
Wird nun ein Skalenton angefordert, So wird diese halbton-differenz auf den Skalenschritt aufaddiert und dann
wird der "nächste" Ton verwendet, um anhand dessen Differenz die Schritte auf der Skala zu gehen.

Beispiel:

nehmen wir die obere Skala und hängen sie auf b' ein.
Hierdurch wird als nächster Ton c" gewählt mit einer Differenz von +13 halbtönen.
Wenn wir jetzt den Skalenton ^-3 anfordern (was ja einer Differenz von -3 gegenüber dem eingehängten Nullpunkt entspricht)
so wählen wir den "nächsten" Ton c" als Nullpunkt, gehen drei Schritte runter (landen beim f#") und addieren
+13 halbtöne (landen beim g"' als Zielton).

Eine Skala wird in der Oktave wiederholt, wenn das Ende der Skala erreicht ist.

also in der oben angegebenen Skala ist der nächste Ton nach c" der Ton d"' und der Ton unter d" der Ton c'
Diese Variante erlaubt es, auch Skalen zu definieren, die partiell nicht oktavrein sind, z.B.

  d#', f',g',a#', c#', d",e",f#",g",a",c"

Hier wäre der Ton oberhalb von c" der Ton d#"' und der Ton unterhalb von d#' wäre c. Jedoch besteht zwischen c#' und c"
keine Oktavreinheit, wie auch zwischen a#' und a" etc.
auch vollständige nicht oktavreine Scalen können definiert werden, wenn die relevanten 5-7 Oktaven komplett absolut definiert werden.

Die Form der Definition ermöglicht, oktavreine Skalen mit minimalem aufwand zu definieren, sie ermöglicht aber auch
die definition von Skalen mit absoluten Ankerpunkten, die nicht oder nur partiell Oktavrein sind.
Die Transposition beim einhängen erfolgt so immer als minimale Verschiebung gegenüber absoluten Ankerpunkten, so dass
der Charakter (etwa bei Instrumenten) maximal gewahrt wird.

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

func (t *ScaleDefinition) Parse(name string, def string) error {
	t.name = name
	t.Original = def
	t.zeroIdx = -1

	notes := strings.Split(def, ",")

	t.definition = make([]int, len(notes))
	t.notes = make([]Note, len(notes))

	var cn = &Note{}
	_ = cn.Parse("a\"", 0)

	t.center = int(cn.ToMIDI())

	for idx, nt := range notes {

		pn := &Note{}

		err := pn.Parse(nt, 0)

		if err != nil {
			return fmt.Errorf("invalid note in scale definition %#v at position %v: %s", t.name, idx, nt)
		}

		mn := int(pn.ToMIDI())
		t.definition[idx] = mn - t.center
		t.notes[idx] = *pn

		if mn-t.center == 0 {
			t.zeroIdx = idx
		}
	}

	if t.zeroIdx == -1 {
		return fmt.Errorf("a\" is missing inside the definition of scale %q", name)
	}

	t.numSteps = len(t.notes)
	max := t.definition[t.numSteps-1]
	min := t.definition[0]
	t.span = max - min + 1

	//fmt.Printf("parsed pattern definition: %#v\n", t)

	return nil

}
