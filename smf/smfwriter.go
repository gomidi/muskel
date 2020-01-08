package smf

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/sketch"
)

type smfWriter struct {
	fileGroup string
	lastNum   uint8
	lastDenom uint8
	score     *score.Score
	delay     *delayWriter
	*mid.SMFWriter
}

func newWriter(s *score.Score, filegroup string) *smfWriter {
	return &smfWriter{score: s, fileGroup: filegroup}
}

func (sw *smfWriter) write(wr smf.Writer) error {
	sw.delay = &delayWriter{Writer: wr}
	sw.SMFWriter = mid.NewSMFWriter(sw.delay)

	err := sw.writeFirstTrack()

	if err != nil {
		return err
	}

	err = sw.writeTempoTrack()

	if err != nil {
		return err
	}

	var tracks []string
	for track := range sw.score.Unrolled {
		tracks = append(tracks, track)
	}

	sort.Strings(tracks)

	//for trackName, events := range sw.score.Unrolled {
	for _, trackName := range tracks {
		events := sw.score.Unrolled[trackName]
		_track, err := sw.score.GetTrack(trackName)
		if err != nil {
			return fmt.Errorf("can't find track %q", trackName)
		}

		// TODO add the params

		if sw.fileGroup != "*" && _track.FileGroup != sw.fileGroup {
			continue
		}

		// ignore tracks without a MIDIchannel
		if _track.MIDIChannel < 0 {
			continue
		}

		sw.delay.firstDeltaSet = false
		sw.delay.track = trackName
		sw.delay.setDelay(_track.Delay[0], _track.Delay[1])
		sw.delay.setStraight()

		tw := newTrackWriter(sw, events, _track)
		tw.writeIntro()
		tw.writeUnrolled()
	}

	return nil
}

func (sw *smfWriter) isStartOfPart(b *sketch.Bar) string {
	for k, v := range sw.score.Parts {
		if v[0] == b.Position {
			return k
		}
	}
	return ""
}

// writeFirstTrack writes the first track with time signature and tempo changes
func (sw *smfWriter) writeFirstTrack() error {

	var bf strings.Builder

	for k, v := range sw.score.Properties() {
		switch strings.ToLower(k) {
		case "copyright":
			fmt.Fprintf(&bf, "(c) %s\n", v)
		default:
			fmt.Fprintf(&bf, "%s: %s\n", k, v)
		}
	}
	sw.Copyright(bf.String())

	num := uint8(4)
	denom := uint8(4)

	var lastBar int

	for _, b := range sw.score.Bars {
		//part := sw.isStartOfPart(b)

		if b.TimeSigChange[0] > 0 || b.Part != "" || b.Comment != "" {
			sw.Forward(0, uint32(b.No-lastBar)*uint32(num), uint32(denom))
			lastBar = b.No
		}

		if b.TimeSigChange[0] > 0 {
			num = b.TimeSigChange[0]
			denom = b.TimeSigChange[1]
			sw.Meter(num, denom)
		}

		//fmt.Printf("write meter %v/%v\n", num, denom)

		var markers []string

		if b.Part != "" {
			markers = append(markers, b.Part)
			//sw.Cuepoint(part)
			//sw.Marker(part)
		}
		if b.Comment != "" {
			markers = append(markers, "// "+b.Comment)
		}

		if len(markers) > 0 {
			marker := strings.Join(markers, " ")
			sw.Cuepoint(marker)
			sw.Marker(marker)
		}
	}

	sw.lastNum = num
	sw.lastDenom = denom

	//p.wr.Forward(0, uint32(p.score.Bars[len(p.score.Bars)-1].BarNo-lastBar)*uint32(num), uint32(denom))
	//p.wr.Meter(num, denom)

	//return nil
	return sw.writeEndOfTrack(lastBar)
}

func (sw *smfWriter) writeTempoTrack() error {
	// iw.wr.Program(iw.instr.Name)
	sw.TrackSequenceName("BPM")
	// iw.wr.Instrument(iw.instr.Name)

	var lastBar int
	var startTempoGlissando uint64
	var startTempoGlissandoValue float64
	var inTempoGlissando bool
	var tempoglissandoFunc func(items.SMFWriter, int64, float64, func(float64))

	for _, b := range sw.score.Bars {

		if b.TempoChange > 0 {
			sw.Forward(uint32(b.No-lastBar), 0, 0)
			lastBar = b.No

			if inTempoGlissando {
				distance := int64(math.Round(float64(sw.Position()+uint64(sw.Delta())-startTempoGlissando) / float64(sw.Ticks16th())))
				diff := b.TempoChange - startTempoGlissandoValue

				tempoglissandoFunc(sw, distance, diff, func(vl float64) {
					vll := vl + startTempoGlissandoValue
					sw.TempoBPM(vll)
					// wr.Text(fmt.Sprintf("%0.2f BPM", vll))
				})
				inTempoGlissando = false
			}

			sw.TempoBPM(b.TempoChange)
			// wr.Text(fmt.Sprintf("%0.2f BPM", b.tempoChange))

			if len(b.Tilde) > 0 {
				startTempoGlissando = sw.Position()
				startTempoGlissandoValue = b.TempoChange
				inTempoGlissando = true
				tempoglissandoFunc = items.LinearTempoChange
				if b.Tilde == "~~" {
					tempoglissandoFunc = items.ExponentialTempoChange
				}
				sw.BackupTimeline()
			}
			//fmt.Printf("write tempo change %v\n", b.tempoChange)
		}
	}

	return sw.writeEndOfTrack(lastBar)
}

func (sw *smfWriter) writeEndOfTrack(lastBarWritten int) error {
	if len(sw.score.Bars) == 0 {
		return fmt.Errorf("no bars found")
	}
	lastBar := sw.score.Bars[len(sw.score.Bars)-1].No
	sw.Forward(uint32(lastBar-lastBarWritten), uint32(sw.lastNum), uint32(sw.lastDenom))
	sw.EndOfTrack()
	return nil
}
