package muskel

import (
	"math"

	"gitlab.com/gomidi/midi/mid"
)

type SMFWriter struct {
	score     *Score
	fileGroup string
	wr        *mid.SMFWriter
}

func NewSMFWriter(s *Score, filegroup string) *SMFWriter {
	return &SMFWriter{score: s, fileGroup: filegroup}
}

func (p *SMFWriter) Write(wr *mid.SMFWriter) error {
	// fmt.Printf("writing score %q is unrolled: %v\n", p.score.FileName, p.score.isUnrolled)
	err := p.writeFirstTrack(wr)

	if err != nil {
		return err
	}

	wr.EndOfTrack()
	err = p.writeTempoTrack(wr)

	if err != nil {
		return err
	}

	for i, instr := range p.score.Instruments {
		if p.fileGroup != "*" && instr.FileGroup != p.fileGroup {
			continue
		}

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

func linearTempoChange(wr *mid.SMFWriter, distance int64, diff float64, callback func(val float64)) {

	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance * step
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance * distance = note0 + noteTarget - note0 = noteTarget
	// y             f(x)        =  a     +      m                      * x
	// m                         =  (noteTarget-note0)/distance

	m := calcAdd(distance, diff)
	//fmt.Printf("linearGlissando: m = %0.5f\n", m)
	wr.RestoreTimeline()
	//var pb int16

	for step := int64(1); step <= distance; step++ {
		//iw.wr.Forward(0, 1, 32)
		wr.Forward(0, 1, 16)
		callback(m * float64(step))
		//pb = halfTonesToPitchbend(m*float64(step), iw.instr.PitchbendRange)
		//iw.wr.Pitchbend(pb)
	}

	//iw.prevPitchbend = pb
}

func exponentialTempoChange(wr *mid.SMFWriter, distance int64, diff float64, callback func(val float64)) {
	// y             f(x)        =  a     +      m                       * x²
	// pitch0      = f(0)        =  note0
	// pitch       = f(step)     =  note0 + (noteTarget-note0)/distance² * step²
	// pitchTarget = f(distance) =  note0 + (noteTarget-note0)/distance² * distance² = note0 + noteTarget - note0 = noteTarget
	// m                         =  (noteTarget-note0)/distance²

	m := calcAdd2(distance, diff)
	wr.RestoreTimeline()
	//var pb int16

	for step := int64(1); step <= distance; step++ {
		//iw.wr.Forward(0, 1, 32)
		wr.Forward(0, 1, 16)
		callback(m * float64(step) * float64(step))
		//pb = halfTonesToPitchbend(m*float64(step)*float64(step), iw.instr.PitchbendRange)
		//iw.wr.Pitchbend(pb)
	}

	//iw.prevPitchbend = pb
}

// writeFirstTrack writes the first track with time signature and tempo changes
func (p *SMFWriter) writeFirstTrack(wr *mid.SMFWriter) error {
	num := uint8(4)
	denom := uint8(4)

	var lastBar int

	for _, b := range p.score.Bars {
		part := p.isStartOfPart(b)

		if b.timeSigChange[0] > 0 || part != "" {
			wr.Forward(0, uint32(b.barNo-lastBar)*uint32(num), uint32(denom))
			lastBar = b.barNo
		}

		if b.timeSigChange[0] > 0 {
			num = b.timeSigChange[0]
			denom = b.timeSigChange[1]
			wr.Meter(num, denom)
		}
		//fmt.Printf("write meter %v/%v\n", num, denom)

		if part != "" {
			wr.Cuepoint(part)
			wr.Marker(part)
		}
	}

	wr.Forward(0, uint32(p.score.Bars[len(p.score.Bars)-1].barNo-lastBar)*uint32(num), uint32(denom))
	wr.Meter(num, denom)

	return nil
}

func (p *SMFWriter) writeTempoTrack(wr *mid.SMFWriter) error {
	// iw.wr.Program(iw.instr.Name)
	wr.TrackSequenceName("BPM")
	// iw.wr.Instrument(iw.instr.Name)

	var lastBar int
	var startTempoGlissando uint64
	var startTempoGlissandoValue float64
	var inTempoGlissando bool
	var tempoglissandoFunc func(*mid.SMFWriter, int64, float64, func(float64))

	for _, b := range p.score.Bars {

		if b.tempoChange > 0 {
			wr.Forward(uint32(b.barNo-lastBar), 0, 0)
			lastBar = b.barNo

			if inTempoGlissando {
				distance := int64(math.Round(float64(wr.Position()+uint64(wr.Delta())-startTempoGlissando) / float64(wr.Ticks16th())))
				diff := b.tempoChange - startTempoGlissandoValue

				tempoglissandoFunc(wr, distance, diff, func(vl float64) {
					vll := vl + startTempoGlissandoValue
					wr.TempoBPM(vll)
					// wr.Text(fmt.Sprintf("%0.2f BPM", vll))
				})
				inTempoGlissando = false
			}

			wr.TempoBPM(b.tempoChange)
			// wr.Text(fmt.Sprintf("%0.2f BPM", b.tempoChange))

			if len(b.tilde) > 0 {
				startTempoGlissando = wr.Position()
				startTempoGlissandoValue = b.tempoChange
				inTempoGlissando = true
				tempoglissandoFunc = linearTempoChange
				if b.tilde == "~~" {
					tempoglissandoFunc = exponentialTempoChange
				}
				wr.BackupTimeline()
			}
			//fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
	}

	return nil
}
