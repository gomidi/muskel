package smf

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/track"
)

type SMF struct {
	fileGroup string
	score     *score.Score
	writer    *Writer
	midiTrack MidiTrack
}

func New(s *score.Score, filegroup string, wr *Writer) *SMF {
	return &SMF{score: s, fileGroup: filegroup, writer: wr}
}

func (s *SMF) write() error {
	var evts events
	meter_evts, err := s.meterTrack()

	if err != nil {
		return err
	}

	tempo_evts, err := s.tempoTrack()

	if err != nil {
		return err
	}

	evts = append(evts, meter_evts...)
	evts = append(evts, tempo_evts...)

	sort.Sort(evts)

	err = s.writer.writeTrack(0, -1, evts)

	if err != nil {
		return err
	}

	var tracks track.Tracks
	for trackName := range s.score.Unrolled {
		tr, err := s.score.GetTrack(trackName)
		if err != nil {
			return fmt.Errorf("can't get the track %q", trackName)
		}
		tracks = append(tracks, tr)
	}

	sort.Sort(tracks)

	for _, track := range tracks {

		if s.fileGroup != "*" && track.FileGroup != s.fileGroup {
			continue
		}

		// ignore tracks without a MIDIchannel
		if track.MIDIChannel < 0 {
			continue
		}

		t := s.newTrack()
		evts = t.trackIntroEvents(track)
		scevts, err := t.trackScoreEvents(track, s.score.Unrolled[track.Name])
		if err != nil && err != smf.ErrFinished {
			return fmt.Errorf("could not write track %q: %v", track.Name, err)
		}
		sort.Sort(scevts)
		evts = append(evts, scevts...)
		endPos := s.posToTicks(track.EndPos)
		err = s.writer.writeTrack(uint8(track.MIDIChannel), int(endPos), evts)

		if err != nil && err != smf.ErrFinished {
			return fmt.Errorf("could not write track %q: %v", track.Name, err)
		}
	}

	return nil
}

func (s *SMF) newTrack() *Track {
	return &Track{
		smf: s,
	}
}

func (s *SMF) isStartOfPart(b *items.Bar) string {
	for k, v := range s.score.Parts {
		if v[0] == b.Position {
			return k
		}
	}
	return ""
}

func (s *SMF) posToTicks(pos uint) uint {
	return (pos * uint(s.writer.ticks) / 8)
}

var regNumBarComment = regexp.MustCompile("^#[0-9]+$")

func (s *SMF) meterTrack() (evts []*event, err error) {

	var bf strings.Builder

	for k, v := range s.score.Properties() {
		switch strings.ToLower(k) {
		case "copyright":
			fmt.Fprintf(&bf, "(c) %s\n", v)
		default:
			fmt.Fprintf(&bf, "%s: %s\n", k, v)
		}
	}

	evts = append(evts, &event{message: smf.MetaCopyright(bf.String())})

	num := uint8(4)
	denom := uint8(4)

	for i, b := range s.score.Bars {
		if b.TimeSigChange[0] > 0 {
			if b.TimeSigChange[0] != num || b.TimeSigChange[1] != denom || i == 0 {
				num = b.TimeSigChange[0]
				denom = b.TimeSigChange[1]
				evts = append(evts, &event{position: s.posToTicks(b.Position), message: smf.MetaMeter(num, denom)})
			}
		}

		var markers []string

		if b.Part != "" {
			markers = append(markers, b.Part)
		}

		comment := strings.TrimSpace(b.Comment)

		if comment != "" && !regNumBarComment.MatchString(comment) {
			markers = append(markers, "// "+comment)
		}

		if len(markers) > 0 {
			marker := strings.Join(markers, " ")
			evts = append(evts, &event{position: s.posToTicks(b.Position), message: smf.MetaCuepoint(marker)})
			evts = append(evts, &event{position: s.posToTicks(b.Position), message: smf.MetaMarker(marker)})
		}
	}
	return
}

func (s *SMF) tempoTrack() (evts []*event, err error) {
	evts = append(evts, &event{position: 0, message: smf.MetaTrackSequenceName("BPM")})

	var glissStartPos uint
	var glissStartBPM float64
	var inGliss bool
	var glissFn func(int64, float64, func(uint, float64))
	lastTempo := float64(120)

	for i, b := range s.score.Bars {

		if b.TempoChange > 0 {
			if inGliss {
				pos := uint(s.writer.ticks)/8*b.Position - glissStartPos

				//distance := int64(math.Round(float64(b.Position-glissStartPos) / float64(s.midiTrack.GlideResolution(s.writer.ticks))))
				distance := int64(math.Round(float64(pos) / float64(s.midiTrack.GlideResolution(s.writer.ticks))))
				diff := b.TempoChange - glissStartBPM
				//fmt.Printf("tempo glissando: \n\tpos: %v distance: %v, diff: %v ticks: %v\n", pos, distance, diff, float64(s.midiTrack.GlideResolution(s.writer.ticks)))

				glissFn(distance, diff, func(step uint, bpm float64) {
					evts = append(evts,
						&event{
							//uint(t.midiTrack.noteGlide.startPosition + uint64(t.smf.writer.ticks/16)*uint64(step)
							//position: glissStartPos + s.posToTicks(step*2),
							position: uint(glissStartPos + uint(s.writer.ticks/16)*uint(step)),
							message:  smf.MetaTempo(bpm + glissStartBPM),
						})
				})
				inGliss = false
			}

			if b.TempoChange != lastTempo || i == 0 {
				evts = append(evts, &event{position: s.posToTicks(b.Position), message: smf.MetaTempo(b.TempoChange)})
				lastTempo = b.TempoChange
			}

			if len(b.Tilde) > 0 {
				glissStartPos = b.Position
				glissStartBPM = b.TempoChange
				inGliss = true
				glissFn = items.LinearTempoChange
				if b.Tilde == "~~" {
					glissFn = items.ExponentialTempoChange
				}
			}
		}

		if len(b.InnerTempoChanges) > 0 {
			sortedTc := sortInternalTempoChanges(b.InnerTempoChanges)

			for _, tc := range sortedTc {
				evts = append(evts, &event{position: s.posToTicks(b.Position + tc.relPos), message: smf.MetaTempo(tc.bpm)})
			}
		}
	}
	return
}

type internalTempoChange struct {
	relPos uint
	bpm    float64
}

type internalTempoChanges []internalTempoChange

func (i internalTempoChanges) Swap(a, b int) {
	i[a], i[b] = i[b], i[a]
}

func (i internalTempoChanges) Less(a, b int) bool {
	return i[a].relPos < i[b].relPos
}

func (i internalTempoChanges) Len() int {
	return len(i)
}

func sortInternalTempoChanges(tc map[uint]float64) internalTempoChanges {
	var sorted internalTempoChanges

	for pos, t := range tc {
		sorted = append(sorted, internalTempoChange{relPos: pos, bpm: t})
	}

	sort.Sort(sorted)

	return sorted
}
