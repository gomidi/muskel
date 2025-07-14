package lilypond

import (
	"gitlab.com/gomidi/lilypond"
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

	for _, b := range s.bars {

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

		if b.timeSignatureChange {
			//	fmt.Printf("bar %v: %v/%v\n", b.no, b.num, b.denom)
			v.Voice.Add(lilypond.TimeSignature{uint(b.num), uint(b.denom)})
		}

		if remainingFromNote > 0 {
			remainingFromNote = remainingNote.writeToVoice(uint(remainingInBar), v.Voice)
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

			remainingFromNote = n.writeToVoice(uint(remainingInBar), v.Voice)
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
