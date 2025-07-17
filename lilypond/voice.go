package lilypond

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/muskel/smf"
)

type voice struct {
	*lilypond.Voice
	track          *Track
	no             int
	openNotes      [128]bool
	notePositions  [128]smf.TicksAbsPos
	noteVelocities [128]uint8
	notes          map[uint][]note // map bar numbers to notes
}

func (v *voice) registerAllMyNotes(s *Score, includeSectionLabels bool) {
	//fmt.Printf("===================TRACK %v VOICE: %v===================\n", v.track.name, v.no)
	/*
		sopranos.NewVoice().SetName("sopranos").Add(ClefTreble(0), global,
			Rest().Quarter(), D(2).Half(), A(1).Quarter(), BarChange,
			D(2).QuarterDot(1), D(2).Eightth(), A(1).Half(), BarChange,
			C(2).Is().Quarter(), D(2), C(2).Is().Half(), BarChange,
		).NewLyrics().Add(words)
	*/

	var remainingFromNote uint
	var remainingNote note
	var currentBarFlatKey bool = false

	for _, b := range s.bars {

		if b.key != nil {
			currentBarFlatKey = b.key.IsFlat
		}

		var remainingInBar int = int(b.Length32())
		//	fmt.Printf("bar: %v\n remaining: %v\n", b.no, remainingInBar)

		if includeSectionLabels && b.marker != "" {
			//v.Voice.Add(lilypond.SectionLabelRoundedBox(b.marker))
			//v.Voice.Add(lilypond.SectionLabel(b.marker))
			v.Voice.Add(lilypond.SectionLabelBox(b.marker))
			//v.Voice.Add(lilypond.Mark(b.marker))
		}

		if b.no > 0 {
			v.Voice.Add(lilypond.BarChange)
		}

		var currentPosInBar uint

		if b.key != nil {
			var key lilypond.Element

			nt := midi.Note(b.key.Key).Name()
			var lynt *lilypond.Note

			switch strings.ToLower(nt) {
			case "c":
				lynt = lilypond.C(0)
			case "cb":
				if b.key.IsFlat {
					lynt = lilypond.C(0).Es()
				} else {
					lynt = lilypond.B(0).Is()
				}
			case "d":
				lynt = lilypond.D(0)
			case "db":
				if b.key.IsFlat {
					lynt = lilypond.D(0).Es()
				} else {
					lynt = lilypond.C(0).Is()
				}
			case "e":
				lynt = lilypond.E(0)
			case "eb":
				if b.key.IsFlat {
					lynt = lilypond.E(0).Es()
				} else {
					lynt = lilypond.D(0).Is()
				}
			case "f":
				lynt = lilypond.F(0)
			case "g":
				lynt = lilypond.G(0)
			case "gb":
				if b.key.IsFlat {
					lynt = lilypond.G(0).Es()
				} else {
					lynt = lilypond.F(0).Is()
				}
			case "a":
				lynt = lilypond.A(0)
			case "ab":
				if b.key.IsFlat {
					lynt = lilypond.A(0).Es()
				} else {
					lynt = lilypond.G(0).Is()
				}
			case "b":
				lynt = lilypond.B(0)
			case "bb":
				if b.key.IsFlat {
					lynt = lilypond.B(0).Es()
				} else {
					lynt = lilypond.A(0).Is()
				}
			default:
				panic(fmt.Sprintf("unknown note: %q", strings.ToLower(nt)))
			}

			switch {
			case b.key.IsMajor:
				key = lilypond.KeyMajor(lynt)
			case !b.key.IsMajor:
				key = lilypond.KeyMinor(lynt)
			default:
				panic("unreachable")
			}

			//		b.key.String()

			//fmt.Printf("key: major: %v flat: %v note: %q lilypond: %q\n", b.key.IsMajor, b.key.IsFlat, nt, key.String())

			v.Voice.Add(key)

			//	lilypond.KeyMajor()
		}

		if b.timeSignatureChange {
			//	fmt.Printf("bar %v: %v/%v\n", b.no, b.num, b.denom)
			v.Voice.Add(lilypond.TimeSignature{uint(b.num), uint(b.denom)})
		}

		if remainingFromNote > 0 {
			remainingFromNote = remainingNote.writeToVoice(uint(remainingInBar), v.Voice, currentBarFlatKey)
			currentPosInBar = remainingNote.length32
			remainingInBar -= int(remainingNote.length32)
			if remainingFromNote > 0 {
				remainingNote = remainingNote.dup()
				remainingNote.length32 = remainingFromNote
				remainingNote.start32thOfBar = 0
				remainingNote.barNo = b.no + 1
				remainingNote.velocity = 64
				remainingNote.accent = false
				continue
			}
		}

		//	fmt.Printf("currentPosInBar A: %v\n", currentPosInBar)
		//	fmt.Printf("remaining A: %v\n", remainingInBar)

		for _, n := range v.notes[b.no] {
			//fmt.Printf("currentPosInBar: %v n.start32thOfBar: %v\n", currentPosInBar, n.start32thOfBar)

			restTime := int(n.start32thOfBar) - int(currentPosInBar)

			if restTime > 0 {
				var rn note
				//		fmt.Printf("resttime: %v\n", restTime)
				rn.addtoVoice(lilypond.Rest(), uint(restTime), false, v.Voice)
				currentPosInBar += uint(restTime)
				remainingInBar -= int(restTime)
			}

			//	fmt.Printf("currentPosInBar B: %v\n", currentPosInBar)
			//		fmt.Printf("remaining B: %v\n", remainingInBar)

			remainingFromNote = n.writeToVoice(uint(remainingInBar), v.Voice, currentBarFlatKey)
			remainingInBar -= int(n.length32)

			if remainingInBar < 0 {
				remainingInBar = 0
				currentPosInBar = b.Length32()
			} else {
				currentPosInBar += n.length32
			}

			//	fmt.Printf("currentPosInBar C: %v\n", currentPosInBar)
			//	fmt.Printf("remaining C: %v\n", remainingInBar)

			if remainingFromNote > 0 {
				remainingNote = n.dup()
				remainingNote.length32 = remainingFromNote
				remainingNote.start32thOfBar = 0
				remainingNote.barNo = b.no + 1
				remainingNote.velocity = 64
				remainingNote.accent = false
				break
			}

			if remainingInBar <= 0 {
				break
			}
		}

		//		fmt.Printf("currentPosInBar D: %v\n", currentPosInBar)
		//		fmt.Printf("remaining D: %v\n", remainingInBar)

		if remainingInBar > 0 {
			var rn note
			//		fmt.Printf("resttime: %v\n", remainingInBar)
			rn.addtoVoice(lilypond.Rest(), uint(remainingInBar), false, v.Voice)
			//currentPosInBar += uint(remainingInBar)
			//remainingInBar = 0
		}

	}

	v.Voice.Add(lilypond.BarChange)
}

func (t *voice) hasNote() bool {
	for _, v := range t.openNotes {
		if v {
			return true
		}
	}

	return false
}
