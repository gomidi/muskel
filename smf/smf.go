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

func (s *SMF) ConvertTrack(track *track.Track) (tr smf.Track, endPos TicksAbsPos, err error) {
	t := s.newTrack()
	evts := t.trackIntroEvents(track)
	scevts, err := t.trackScoreEvents(track, s.score.Unrolled[track.Name])
	if err != nil && err != smf.ErrFinished {
		err = fmt.Errorf("could not write track %q: %v", track.Name, err)
		return
	}
	sort.Sort(scevts)
	evts = append(evts, scevts...)
	endPos = s.posToTicks(track.EndPos)
	tr, err = s.writeEventsToTrack(uint8(track.MIDIChannel), int(endPos), evts)
	return
}

func (s *SMF) writeEventsToTrack(ch uint8, endPos int, evts Events) (tr smf.Track, err error) {
	return s.writer.writeTrack(ch, endPos, evts)
}

func (s *SMF) addTrack(tr smf.Track) error {
	err := s.writer.SMF.Add(tr)
	if err == smf.ErrFinished {
		return nil
	}
	s.writer.currentTrack++
	return err
}

func (s *SMF) addConvertedTrack(track *track.Track) (err error) {
	tr, _, err := s.ConvertTrack(track)

	if err != nil {
		return err
	}

	return s.addTrack(tr)
}

func (s *SMF) addTrackFromEvents(ch uint8, endPos int, evts Events) (err error) {
	tr, err := s.writeEventsToTrack(ch, endPos, evts)
	if err != nil {
		return err
	}
	return s.addTrack(tr)
}

func (s *SMF) write() error {
	var evts Events
	meter_evts, err := s.MeterTrack()

	if err != nil {
		return err
	}

	tempo_evts, err := s.TempoTrack()

	if err != nil {
		return err
	}

	evts = append(evts, meter_evts...)
	evts = append(evts, tempo_evts...)

	sort.Sort(evts)

	err = s.addTrackFromEvents(0, -1, evts)

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

		err = s.addConvertedTrack(track)

		/*
			t := s.newTrack()
			evts = t.trackIntroEvents(track)
			scevts, err := t.trackScoreEvents(track, s.score.Unrolled[track.Name])
			if err != nil && err != smf.ErrFinished {
				return fmt.Errorf("could not write track %q: %v", track.Name, err)
			}
			sort.Sort(scevts)
			evts = append(evts, scevts...)
			endPos := s.posToTicks(track.EndPos)
			err = s.writeTrack(uint8(track.MIDIChannel), int(endPos), evts)
		*/

		if err != nil && err != smf.ErrFinished {
			return fmt.Errorf("could not write track %q: %v", track.Name, err)
		}
	}

	_, err = s.writer.SMF.WriteTo(s.writer.targetWR)
	return err
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

func (s *SMF) TicksPerQN() uint {
	return uint(s.writer.ticks)
}

func (s *SMF) posToTicks(pos uint) TicksAbsPos {
	return TicksAbsPos(pos * s.TicksPerQN() / 8)
}

var regNumBarComment = regexp.MustCompile("^#[0-9]+$")

func (s *SMF) MeterTrack() (evts []*Event, err error) {

	var bf strings.Builder

	for k, v := range s.score.Properties() {
		switch strings.ToLower(k) {
		case "copyright":
			fmt.Fprintf(&bf, "(c) %s\n", v)
		default:
			fmt.Fprintf(&bf, "%s: %s\n", k, v)
		}
	}

	evts = append(evts, &Event{Message: smf.MetaCopyright(bf.String())})

	num := uint8(4)
	denom := uint8(4)

	scale := `major_c`

	for i, b := range s.score.Bars {

		nt, aug, _ := items.KeyToNote(b.Scale.BaseNote)
		_scale := fmt.Sprintf("%s_%s%s", b.Scale.Name, nt, aug)

		if _scale != scale {
			evts = append(evts, &Event{Position: s.posToTicks(b.Position), Message: smf.MetaMeter(num, denom)})

			var key smf.Message
			var setKey = true

			switch b.Scale.Name {
			case "major":
				switch nt + aug {
				case "c":
					key = smf.CMaj()
				case "c#", "db":
					key = smf.DbMaj()
				case "d":
					key = smf.DMaj()
				case "d#", "eb":
					key = smf.EbMaj()
				case "e":
					key = smf.EMaj()
				case "f":
					key = smf.FMaj()
				case "f#":
					key = smf.FsharpMaj()
				case "gb":
					key = smf.GbMaj()
				case "g":
					key = smf.GMaj()
				case "g#", "ab":
					key = smf.AbMaj()
				case "a":
					key = smf.AMaj()
				case "a#", "bb":
					key = smf.AbMaj()
				default:
					setKey = false
				}
			case "minor":
				switch nt + aug {
				case "c":
					key = smf.CMin()
				case "c#", "db":
					key = smf.CsharpMin()
				case "d":
					key = smf.DMin()
				case "d#":
					key = smf.DsharpMin()
				case "eb":
					key = smf.EbMin()
				case "e":
					key = smf.EMin()
				case "f":
					key = smf.FMin()
				case "f#", "gb":
					key = smf.FsharpMin()
				case "g":
					key = smf.GMin()
				case "g#", "ab":
					key = smf.GsharpMin()
				case "a":
					key = smf.AMin()
				case "a#", "bb":
					key = smf.BbMin()
				default:
					setKey = false
				}
			default:
				setKey = false
			}

			if setKey {
				evts = append(evts, &Event{Position: s.posToTicks(b.Position), Message: key})
			}

			scale = _scale
		}

		if b.TimeSigChange[0] > 0 {
			if b.TimeSigChange[0] != num || b.TimeSigChange[1] != denom || i == 0 {
				num = b.TimeSigChange[0]
				denom = b.TimeSigChange[1]
				evts = append(evts, &Event{Position: s.posToTicks(b.Position), Message: smf.MetaMeter(num, denom)})
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
			evts = append(evts, &Event{Position: s.posToTicks(b.Position), Message: smf.MetaCuepoint(marker)})
			evts = append(evts, &Event{Position: s.posToTicks(b.Position), Message: smf.MetaMarker(marker)})
		}
	}
	return
}

func (s *SMF) TempoTrack() (evts []*Event, err error) {
	evts = append(evts, &Event{Position: 0, Message: smf.MetaTrackSequenceName("BPM")})

	var glissStartPos uint
	var glissStartBPM float64
	var inGliss bool
	var glissFn func(int64, float64, func(uint, float64))
	lastTempo := float64(120)

	for i, b := range s.score.Bars {

		if b.TempoChange > 0 {
			if inGliss {
				pos := uint(s.writer.ticks)/8*b.Position - glissStartPos

				distance := int64(math.Round(float64(pos) / float64(s.midiTrack.GlideResolution(s.writer.ticks))))
				diff := b.TempoChange - glissStartBPM

				glissFn(distance, diff, func(step uint, bpm float64) {
					evts = append(evts,
						&Event{
							Position: TicksAbsPos(uint(glissStartPos + uint(s.writer.ticks/16)*uint(step))),
							Message:  smf.MetaTempo(bpm + glissStartBPM),
						})
				})
				inGliss = false
			}

			if b.TempoChange != lastTempo || i == 0 {
				evts = append(evts, &Event{Position: s.posToTicks(b.Position), Message: smf.MetaTempo(b.TempoChange)})
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
				evts = append(evts, &Event{Position: s.posToTicks(b.Position + tc.relPos), Message: smf.MetaTempo(tc.bpm)})
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
