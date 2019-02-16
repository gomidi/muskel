package muskel

import (
	"gitlab.com/gomidi/midi/mid"
)

type SMFWriter struct {
	score        *Score
	unrolledbars []*Bar
	wr           *mid.SMFWriter
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

	for _, instr := range p.score.Instruments {
		// ignore instruments without a MIDIchannel
		if instr.MIDIChannel < 0 {
			continue
		}

		//		fmt.Println("EOT")
		wr.EndOfTrack()
		iw := newInstrumentSMFWriter(p, wr, instr)
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

	for i, b := range p.unrolledbars {
		/*
			b.positions = make([]uint8, len(b.originalPositions))
			var oldPos string

			for pi, pos := range b.originalPositions {
				oldPos, b.positions[pi] = positionTo32th(oldPos, pos)
			}
		*/

		//fmt.Printf("bar %v: %#v\n", i, b)

		if i > 0 {
			//fmt.Printf("[firstrack] Forward(1, 0, 0)\n")
			wr.Forward(1, 0, 0)
			/*
				fmt.Printf("Forward(0, %v, %v)\n", uint32(num), uint32(denom))
				wr.Forward(0, uint32(num), uint32(denom))
			*/
		}

		if b.timeSigChange[0] > 0 {
			num = b.timeSigChange[0]
			denom = b.timeSigChange[1]
		}
		//fmt.Printf("write meter %v/%v\n", num, denom)

		if b.tempoChange > 0 {
			wr.TempoBPM(b.tempoChange)
			//fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
		wr.Meter(num, denom)
		//wr.Forward(1, 0, 0)
	}

	return nil
}
