package muskel

import (
	"fmt"

	"gitlab.com/gomidi/midi/mid"
)

type SMFWriter struct {
	score *Score
	//unrolledbars []*Bar
	wr *mid.SMFWriter
}

func NewSMFWriter(s *Score) *SMFWriter {
	return &SMFWriter{score: s}
}

func (p *SMFWriter) Write(wr *mid.SMFWriter) error {

	/*
		err := p.unrollInstruments()
		if err != nil {
			return err
		}
	*/
	err := p.writeFirstTrack(wr)

	if err != nil {
		return err
	}

	for i, instr := range p.score.Instruments {
		// ignore instruments without a MIDIchannel
		if instr.MIDIChannel < 0 {
			continue
		}

		//		fmt.Println("EOT")
		wr.EndOfTrack()
		iw := newInstrumentSMFWriter(p, wr, instr)
		fmt.Printf("writing MIDI for col: %v, instr: %q\n", i, instr.Name)
		iw.writeIntro()
		err = iw.writeUnrolled()
		//err = iw.writeTrack()
		//		err = p.writeInstrumentTrack(wr, instr)
		if err != nil {
			return err
		}
	}

	//	fmt.Println("EOT")
	wr.EndOfTrack()
	return nil
}

// writeFirstTrack writes the first track with time signature and tempo changes
func (p *SMFWriter) writeFirstTrack(wr *mid.SMFWriter) error {
	num := uint8(4)
	denom := uint8(4)

	var lastBar int

	for _, b := range p.score.Bars {
		/*
			b.positions = make([]uint8, len(b.originalPositions))
			var oldPos string

			for pi, pos := range b.originalPositions {
				oldPos, b.positions[pi] = positionTo32th(oldPos, pos)
			}
		*/

		//fmt.Printf("bar %v: %#v\n", i, b)

		//		if i > 0 {
		//fmt.Printf("[firstrack] Forward(1, 0, 0)\n")
		//			wr.Forward(1, 0, 0)
		/*
			fmt.Printf("Forward(0, %v, %v)\n", uint32(num), uint32(denom))
			wr.Forward(0, uint32(num), uint32(denom))
		*/
		//		}

		if b.timeSigChange[0] > 0 || b.tempoChange > 0 {
			wr.Forward(uint32(b.barNo-lastBar), 0, 0)
			lastBar = b.barNo
		}

		if b.timeSigChange[0] > 0 {
			num = b.timeSigChange[0]
			denom = b.timeSigChange[1]
			wr.Meter(num, denom)
		}
		//fmt.Printf("write meter %v/%v\n", num, denom)

		if b.tempoChange > 0 {
			wr.TempoBPM(b.tempoChange)
			//fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
		//wr.Forward(1, 0, 0)
	}

	return nil
}
