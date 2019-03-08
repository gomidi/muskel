package muskel

import (
	"gitlab.com/gomidi/midi/mid"
)

type SMFWriter struct {
	score *Score
	wr    *mid.SMFWriter
}

func NewSMFWriter(s *Score) *SMFWriter {
	return &SMFWriter{score: s}
}

func (p *SMFWriter) Write(wr *mid.SMFWriter) error {
	// fmt.Printf("writing score %q is unrolled: %v\n", p.score.FileName, p.score.isUnrolled)
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
		_ = i
		//fmt.Printf("writing MIDI for col: %v, instr: %q\n", i, instr.Name)
		iw.writeIntro()
		iw.writeUnrolled()
	}

	//	fmt.Println("EOT")
	wr.EndOfTrack()
	return nil
}

func (p *SMFWriter) isStartOfPart(b *Bar) string {
	for k, v := range p.score.Parts {
		if v[0] == b.barNo {
			return k
		}
	}
	return ""
}

// writeFirstTrack writes the first track with time signature and tempo changes
func (p *SMFWriter) writeFirstTrack(wr *mid.SMFWriter) error {
	num := uint8(4)
	denom := uint8(4)

	var lastBar int

	for _, b := range p.score.Bars {

		part := p.isStartOfPart(b)

		if b.timeSigChange[0] > 0 || b.tempoChange > 0 || part != "" {
			wr.Forward(0, uint32(b.barNo-lastBar)*uint32(num), uint32(denom))
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

		if part != "" {
			wr.Cuepoint(part)
			wr.Marker(part)
		}
	}

	return nil
}
