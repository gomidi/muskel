package lilypond

import (
	"fmt"

	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/midi/v2"
)

type note struct {
	key            uint8
	start32thOfBar uint
	barNo          uint
	length32       uint
	velocity       uint8
}

func (n note) dup() note {
	return n
}

func (n note) getVelocity(nt *lilypond.Note) *lilypond.Note {
	/*
		12 stufen + nix = 13 stufen
		wenn wir 3 zu den 127 hinzuaddieren, sind wir bei 130.
		dann können wir durch 13 teilen und dementsprechend verteilen.
	*/

	vn := (n.velocity + 3) / 10

	switch {
	case vn >= 13:
		return nt.FFFFF()
	case vn >= 12:
		return nt.FFFF()
	case vn >= 11:
		return nt.FFF()
	case vn >= 9:
		return nt.FF()
	case vn >= 8:
		return nt.F()
	case vn >= 7:
		//	return nt.MF()
		return nt
	case vn >= 6:
		return nt
	case vn >= 5:
		//	return nt.MP()
		return nt
	case vn >= 4:
		return nt.P()
	case vn >= 3:
		return nt.PP()
	case vn >= 2:
		return nt.PPP()
	case vn >= 1:
		return nt.PPPP()
	default:
		return nt.PPPPP()
	}
}

func (n note) addtoVoice(nt *lilypond.Note, length32 uint, hasRest bool, v *lilypond.Voice) {

	//	fmt.Printf("len 32: %v\n", length32)
	var rest uint

	switch {
	case length32 >= 32: // wholes
		//		fmt.Println("setting wholes or more")
		no := length32 / 32
		rest = length32 % 32
		switch no {
		case 8:
			nt = nt.Maxima()
		case 7:
			nt = nt.LongaDot(2)
		case 6:
			nt = nt.LongaDot(1)
		case 5:
			nt = nt.Longa()
			rest += 32
		case 4:
			nt = nt.Longa()
		case 3:
			nt = nt.BreveDot(1)
		case 2:
			nt = nt.Breve()
		case 1:
			nt = nt.Whole()
		default:
			panic(fmt.Sprintf("should not happen: %v 32th inside bar. bar is too long", no))
		}

	case length32 >= 16: // halves
		nt = nt.Half()
		//	fmt.Println("setting half")
		rest = length32 % 16
	case length32 >= 8: // QN
		//	fmt.Println("setting QN")
		nt = nt.Quarter()
		rest = length32 % 8
	case length32 >= 4: // 8th
		//	fmt.Println("setting 8th")
		nt = nt.Eightth()
		rest = length32 % 4
	case length32 >= 2: // 16th
		//	fmt.Println("setting 16th")
		nt = nt.Sixteenth()
		rest = length32 % 2
	default:
		//	fmt.Println("setting 32th")
		nt = nt.ThirtySecond()
	}

	if rest > 0 || hasRest {
		nt = nt.Tie()
	}

	//	fmt.Printf("adding: %s\n", nt.String())

	v.Add(nt)

	if rest > 0 {
		//	fmt.Printf("adding rest: %v\n", rest)
		n.addtoVoice(nt, rest, hasRest, v)
	}

}

func (n note) getLilyNote() *lilypond.Note {
	nt := midi.Note(n.key)

	oct := int(nt.Octave()) - 3
	if oct < 0 {
		oct = 0
	}

	switch nt.Name() {
	case "C":
		return lilypond.C(oct)
	case "Db":
		return lilypond.D(oct).Es()
	case "D":
		return lilypond.D(oct)
	case "Eb":
		return lilypond.E(oct).Es()
	case "E":
		return lilypond.E(oct)
	case "F":
		return lilypond.F(oct)
	case "Gb":
		return lilypond.G(oct).Es()
	case "G":
		return lilypond.G(oct)
	case "Ab":
		return lilypond.A(oct).Es()
	case "A":
		return lilypond.A(oct)
	case "Bb":
		return lilypond.B(oct).Es()
	case "B":
		return lilypond.B(oct)
	default:
		panic("unreachable")
	}

}

func (n note) writeToVoice(remaining32thInBar uint, v *lilypond.Voice) (remainingLength32 uint) {
	nt := n.getLilyNote()
	nt = n.getVelocity(nt)
	diff := int(n.length32) - int(remaining32thInBar)
	if diff <= 0 {
		n.addtoVoice(nt, n.length32, false, v)
		return 0
	}

	n.addtoVoice(nt, remaining32thInBar, true, v)
	return uint(diff)
}
