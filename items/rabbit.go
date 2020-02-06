package items

import (
	"fmt"
	"math"
	"strconv"
)

type Rabbit struct {
	note    Note
	number  uint
	factor  float64
	percent float64
}

// params:
// 1. note
// 2. factor
// 3. percent
// 4. number of notes
func (e *Rabbit) Parse(startPos32th uint, params ...string) (err error) {
	if len(params) != 4 {
		return fmt.Errorf("Rabbit must have 4 params")
	}

	var n Note
	err = n.Parse(params[0], 0)

	if err != nil {
		return fmt.Errorf("first parameter must be a note, but is %q", params[0])
	}

	e.note = n

	e.factor, err = strconv.ParseFloat(params[1], 64)

	if err != nil {
		return fmt.Errorf("factor of positions (second parameter) must be a float")
	}

	if e.factor <= 0 {
		return fmt.Errorf("factor of positions (second parameter) must be > 0")
	}

	if len(params[2]) < 2 {
		return fmt.Errorf("invalid percent number (third parameter): %q", params[2])
	}

	if params[2][len(params[2])-1] != '%' {
		return fmt.Errorf("invalid percent number (third parameter): %q (must end with percent sign)", params[2])
	}

	perc := params[2][:len(params[2])-1]
	var pr float64
	pr, err = strconv.ParseFloat(perc, 64)

	if err != nil {
		return fmt.Errorf("percent number (third parameter) must be a float")
	}

	if pr <= 0 {
		return fmt.Errorf("percent number (third parameter) must be > 0")
	}

	pr = pr / 100.0

	if pr > 1 {
		return fmt.Errorf("percent number (third parameter) must be < 100")
	}

	e.percent = pr

	var i int
	i, err = strconv.Atoi(params[3])

	if err != nil {
		return fmt.Errorf("number of notes (fourth parameter) must be an integer")
	}

	if i <= 1 {
		return fmt.Errorf("number of notes (fourth parameter) must be at least 2")
	}

	e.number = uint(i)

	return nil
}

func (r *Rabbit) next(prev uint8) (midinote uint8) {
	mn := math.Round(float64(prev) * r.factor * (1.0 - r.percent))
	if mn < 0 {
		mn = mn * (-1.0)
	}
	midinote = uint8(mn)
	if midinote > 127 {
		midinote = 127
	}
	return
}

func (r *Rabbit) scaleStrings() (s []string) {
	nts := r.notes()

	for _, nt := range nts {
		var n Note
		n.ScaleNote = int8(nt)
		s = append(s, n.String())
	}

	return
}

func (r *Rabbit) Strings() (s []string) {
	if r.note.ScaleNote != 0 {
		return r.scaleStrings()
	}
	nts := r.notes()

	for _, nt := range nts {
		n := keyToNote(nt)
		s = append(s, n.String())
	}

	return
}

func (r *Rabbit) scaleNotes() (nt []uint8) {
	var n uint
	note := uint8(r.note.ScaleNote)
	nt = append(nt, note)
	n++

	for n < r.number {
		nn := r.next(note)
		if nn > 0 {
			note = nn
			nt = append(nt, note)
			n++
		}
	}

	return
}

func (r *Rabbit) notes() (nt []uint8) {
	if r.note.ScaleNote != 0 {
		return r.scaleNotes()
	}
	var n uint
	note := r.note.ToMIDI()
	nt = append(nt, note)
	n++

	for n < r.number {
		nn := r.next(note)
		if nn > 0 {
			note = nn
			nt = append(nt, note)
			n++
		}
	}

	return
}
