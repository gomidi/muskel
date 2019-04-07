package smf

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/unroller"
)

type errors []error

func (e errors) Error() string {
	var bf strings.Builder
	bf.WriteString("The following errors happened:\n")
	for _, err := range e {
		bf.WriteString(err.Error() + "\n")
	}

	return bf.String()
}

type debugLog struct{}

func (debugLog) Printf(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", vals...)
}

func WriteSMFTo(s *score.Score, wr io.Writer, filegroup string, options ...smfwriter.Option) error {
	numTracks := uint16(2) // first track is for time signatures, second track is for tempo changes

	for _, instr := range s.Tracks {
		if instr.MIDIChannel >= 0 && (filegroup == "*" || instr.FileGroup == filegroup) {
			numTracks++
		}
	}

	options = append(
		[]smfwriter.Option{
			smfwriter.TimeFormat(smf.MetricTicks(960)),
			smfwriter.NumTracks(numTracks),
		}, options...)

	if score.DEBUG {
		options = append(options, smfwriter.Debug(debugLog{}))
	}

	sw := newWriter(s, filegroup)

	return sw.Write(smfwriter.New(wr, options...))

}

// WriteFile writes the score to the given SMF file
func WriteFile(s *score.Score, midifile string, options ...smfwriter.Option) (err error) {
	defer func() {
		if !score.DEBUG {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	if !s.IsUnrolled {
		ur, err := unroller.Unroll(s)
		if err != nil {
			return err
		}
		return WriteFile(ur, midifile, options...)
	}

	hasPlaceholder := strings.Index(midifile, "%s") > -1

	if !hasPlaceholder {
		return writeSMFToFile(s, midifile, "*", options...)
	}

	var fileGroups = map[string]string{}
	for _, instr := range s.Tracks {
		fileGroups[instr.FileGroup] = fmt.Sprintf(midifile, instr.FileGroup)
	}

	var errs errors

	for grp, fl := range fileGroups {
		err := writeSMFToFile(s, fl, grp, options...)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func writeSMFToFile(s *score.Score, midifile, filegroup string, options ...smfwriter.Option) error {
	f, err := os.Create(midifile)
	if err != nil {
		return err
	}

	defer f.Close()

	return WriteSMFTo(s, f, filegroup)

	/*

		numTracks := uint16(2) // first track is for time signatures, second track is for tempo changes

		for _, instr := range s.Instruments {
			if instr.MIDIChannel >= 0 && (filegroup == "*" || instr.FileGroup == filegroup) {
				numTracks++
			}
		}

		sw := NewSMFWriter(s, filegroup)

		options = append(
			[]smfwriter.Option{
				smfwriter.TimeFormat(smf.MetricTicks(960)),
			}, options...)

		if DEBUG {
			options = append(options, smfwriter.Debug(debugLog{}))
		}

		return mid.NewSMFFile(midifile, numTracks, sw.Write, options...)
	*/
}

type writer struct {
	score     *score.Score
	fileGroup string
	wr        *mid.SMFWriter
	iw        *instrSMFWriter
}

func newWriter(s *score.Score, filegroup string) *writer {
	return &writer{score: s, fileGroup: filegroup}
}

func (p *writer) Write(wr smf.Writer) error {

	p.iw = &instrSMFWriter{Writer: wr, instrNo: -1}
	p.wr = mid.NewSMFWriter(p.iw)

	// fmt.Printf("writing score %q is unrolled: %v\n", p.score.FileName, p.score.isUnrolled)
	err := p.writeFirstTrack()

	if err != nil {
		return err
	}

	p.wr.EndOfTrack()
	err = p.writeTempoTrack()

	if err != nil {
		return err
	}

	for i, instr := range p.score.Tracks {
		if p.fileGroup != "*" && instr.FileGroup != p.fileGroup {
			continue
		}

		// ignore instruments without a MIDIchannel
		if instr.MIDIChannel < 0 {
			continue
		}

		//		fmt.Println("EOT")
		p.wr.EndOfTrack()
		p.iw.firstDeltaSet = false
		p.iw.instrNo = i
		// fmt.Printf("instr %q  delay %v\n", instr.Name, instr.Delay)
		p.iw.setDelay(instr.Delay[0], instr.Delay[1])
		p.iw.setStraight()

		iw := newTrackSMFWriter(p, p.wr, instr)
		_ = i
		//fmt.Printf("writing MIDI for col: %v, instr: %q\n", i, instr.Name)
		iw.writeIntro()
		iw.writeUnrolled()
	}

	//	fmt.Println("EOT")
	p.wr.EndOfTrack()
	return nil
}

func (p *writer) isStartOfPart(b *score.Bar) string {
	for k, v := range p.score.Parts {
		if v[0] == b.BarNo {
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
func (p *writer) writeFirstTrack() error {

	var bf strings.Builder

	for k, v := range p.score.Meta {
		switch strings.ToLower(k) {
		case "copyright":
			fmt.Fprintf(&bf, "(c) %s\n", v)
		default:
			fmt.Fprintf(&bf, "%s: %s\n", k, v)
		}
	}
	p.wr.Copyright(bf.String())

	num := uint8(4)
	denom := uint8(4)

	var lastBar int

	for _, b := range p.score.Bars {
		part := p.isStartOfPart(b)

		if b.TimeSigChange[0] > 0 || part != "" {
			p.wr.Forward(0, uint32(b.BarNo-lastBar)*uint32(num), uint32(denom))
			lastBar = b.BarNo
		}

		if b.TimeSigChange[0] > 0 {
			num = b.TimeSigChange[0]
			denom = b.TimeSigChange[1]
			p.wr.Meter(num, denom)
		}
		//fmt.Printf("write meter %v/%v\n", num, denom)

		if part != "" {
			p.wr.Cuepoint(part)
			p.wr.Marker(part)
		}
	}

	p.wr.Forward(0, uint32(p.score.Bars[len(p.score.Bars)-1].BarNo-lastBar)*uint32(num), uint32(denom))
	p.wr.Meter(num, denom)

	return nil
}

func (p *writer) writeTempoTrack() error {
	// iw.wr.Program(iw.instr.Name)
	p.wr.TrackSequenceName("BPM")
	// iw.wr.Instrument(iw.instr.Name)

	var lastBar int
	var startTempoGlissando uint64
	var startTempoGlissandoValue float64
	var inTempoGlissando bool
	var tempoglissandoFunc func(*mid.SMFWriter, int64, float64, func(float64))

	for _, b := range p.score.Bars {

		if b.TempoChange > 0 {
			p.wr.Forward(uint32(b.BarNo-lastBar), 0, 0)
			lastBar = b.BarNo

			if inTempoGlissando {
				distance := int64(math.Round(float64(p.wr.Position()+uint64(p.wr.Delta())-startTempoGlissando) / float64(p.wr.Ticks16th())))
				diff := b.TempoChange - startTempoGlissandoValue

				tempoglissandoFunc(p.wr, distance, diff, func(vl float64) {
					vll := vl + startTempoGlissandoValue
					p.wr.TempoBPM(vll)
					// wr.Text(fmt.Sprintf("%0.2f BPM", vll))
				})
				inTempoGlissando = false
			}

			p.wr.TempoBPM(b.TempoChange)
			// wr.Text(fmt.Sprintf("%0.2f BPM", b.tempoChange))

			if len(b.Tilde) > 0 {
				startTempoGlissando = p.wr.Position()
				startTempoGlissandoValue = b.TempoChange
				inTempoGlissando = true
				tempoglissandoFunc = linearTempoChange
				if b.Tilde == "~~" {
					tempoglissandoFunc = exponentialTempoChange
				}
				p.wr.BackupTimeline()
			}
			//fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
	}

	return nil
}
