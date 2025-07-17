package lilypond

import (
	"fmt"
	"sort"
	"strings"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/lilypond"
	midismf "gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
	"gitlab.com/gomidi/muskel/track"
)

type Option func(*Score)

type Score struct {
	score                 *score.Score
	fileGroup             string
	smf                   *smf.SMF
	currentBarNo          uint
	currentBarTick        smf.TicksAbsPos
	lastEvTick            smf.TicksAbsPos
	currentTimeSignature  [2]uint8
	currentTempo          float64
	timeSignatureChanges  [][3]uint                     // 0 = tick 1 num 2 denom
	tempoChanges          []map[smf.TicksAbsPos]float64 // tick => value
	keyChanges            map[uint]midismf.Key          // tick => val
	startingTimeSignature [2]uint8
	ticksPerQN            uint
	startingTempo         float64
	bars                  []bar
	endPos                smf.TicksAbsPos
	tracks                []*Track
	copyright             string
	title                 string
	subtitle              string
	composer              string
	poet                  string
	arranger              string
	lastOpenedNote        uint8
	markers               map[uint]string // absticks => text
}

func newScore(s *score.Score, filegroup string) *Score {
	wrt := smf.NewWriter(nil, 1000, 960)
	sw := smf.New(s, filegroup, wrt)
	//	fmt.Println("after init")
	props := s.Properties()

	newsc := &Score{score: s, fileGroup: filegroup, smf: sw, startingTimeSignature: [2]uint8{4, 4}}
	newsc.markers = map[uint]string{}
	newsc.keyChanges = map[uint]midismf.Key{}
	if len(props) > 0 {
		newsc.composer = props["composer"]
		newsc.title = props["title"]
		newsc.copyright = props["copyright"]
		newsc.poet = props["poet"]
		newsc.arranger = props["arranger"]
		newsc.subtitle = props["subtitle"]
	}

	return newsc
}

func (s *Score) TicksTo32th(ticks uint) uint {
	return (ticks * s.ticksPerQN) / 8
}

func (s *Score) newTrack(name string, midichannel int8, clef string, dynhandling track.LilyPondDynamics) *Track {
	var tr = &Track{}
	tr.lyrics = map[uint]string{} // barnumber to lyrics
	tr.texts = map[smf.TicksAbsPos]string{}
	tr.name = name
	tr.staff = new(lilypond.Staff).SetName(name).With(lilypond.InstrumentName(name))
	tr.clef = clef
	if tr.clef == "" && midichannel == 9 {
		tr.clef = "percussion"
	}
	tr.dynhandling = dynhandling
	tr.addVoice()
	return tr
}

func (s *Score) closeNote(tr *Track, key uint8) {
	v := tr.closeOpenedNote(key)
	startPos := v.notePositions[key]
	length := uint(s.lastEvTick - startPos)
	b := s.barOfPos(startPos)

	var n note
	n.start32thOfBar = uint((startPos-b.position)*8) / s.ticksPerQN
	n.length32 = (length * 8) / s.ticksPerQN
	n.key = key

	switch {
	case tr.dynhandling == track.LilyPondDynHide:
		n.velocity = 64
	case tr.isPercussion() || tr.dynhandling == track.LilyPondDynAccents:
		n.velocity = 64
		if v.noteVelocities[key] > 100 {
			n.accent = true
		}
	default:
		n.velocity = v.noteVelocities[key]
	}

	n.barNo = b.no
	v.notes[b.no] = append(v.notes[b.no], n)

	//fmt.Printf("in track %v bar %v note %v start: %v length32 %v voice: %v\n", track.name, b.no, midi.Note(key).String(), n.start32thOfBar, n.length32, v.no)
}

// 1. we need to calculate the bar, therefor we need all the time-signatures and their bar positions
// 2. we need to calculate the beat/whatever position from the tick (and also round it somehow)
// 3. we should also calculate the key, but this is more complicated and we do it later
func (s *Score) addEventToTrack(track *Track, ev midismf.Event) {

	s.lastEvTick = s.lastEvTick + smf.TicksAbsPos(ev.Delta)

	var ch, key, vel uint8
	var text string
	switch {
	case ev.Message.GetMetaLyric(&text):
		b := s.barOfPos(s.lastEvTick)
		track.lyrics[b.no] = track.lyrics[b.no] + text + " "
	case ev.Message.GetMetaText(&text):
		track.texts[s.lastEvTick] = text
	case ev.Message.GetNoteStart(&ch, &key, &vel):
		if track.isPercussion() && s.lastOpenedNote != 0 {
			s.closeNote(track, s.lastOpenedNote)
		}

		v := track.openNote(key, vel, s.lastEvTick)
		s.lastOpenedNote = key
		_ = v
	case ev.Message.GetNoteEnd(&ch, &key):
		if !track.isPercussion() {
			s.closeNote(track, key)
		}
	}
}

func (s *Score) barOfPos(pos smf.TicksAbsPos) bar {
	var last = s.bars[0]
	for _, b := range s.bars {
		if b.position > pos {
			return last
		}
		last = b
	}

	return last
}

func (s *Score) addBar(b bar) {
	//	fmt.Printf("adding bar %v\n", b.no)
	s.bars = append(s.bars, b)
}

func (s *Score) setMarkesAndKeyChangesToBars() {
	for i, bar := range s.bars {
		text, has := s.markers[bar.position.Uint()]

		if has {
			bar.marker = text
			s.bars[i] = bar
		}

		key, hasKey := s.keyChanges[bar.position.Uint()]

		if hasKey {
			bar.key = &key
			s.bars[i] = bar
		}
	}
}

func (s *Score) setBars() error {

	var num, denom uint8

	num = s.startingTimeSignature[0]
	denom = s.startingTimeSignature[1]
	var barno uint
	var lastPos smf.TicksAbsPos

	var b bar
	b.no = barno
	b.timeSignatureChange = true
	b.position = lastPos
	b.num = num
	b.denom = denom
	b.ticksPerQN = s.ticksPerQN
	//s.bars = append(s.bars, b)
	s.addBar(b)

	//	fmt.Printf("no s.timeSignatureChanges: %v\n", len(s.timeSignatureChanges))

	for _, ts := range s.timeSignatureChanges {
		//	fmt.Printf("time signatature %v\n", j)
		pos, _num, _denom := ts[0], ts[1], ts[2]
		if pos == 0 { // we already checked this via startingTimeSignature in scanMeterChanges
			//	fmt.Println("skip the first")
			continue
		}

		var bnew bar

		bnew.position = smf.TicksAbsPos(pos)
		bnew.num = uint8(_num)
		bnew.denom = uint8(_denom)
		bnew.ticksPerQN = s.ticksPerQN
		bnew.timeSignatureChange = true

		var diffTicks uint
		var missingBars uint

		if b.End() < bnew.position {
			diffTicks = uint(bnew.position - b.End())
			missingBars = diffTicks / b.Length()
		}

		for i := 0; i < int(missingBars); i++ {
			bx := b.dup()
			bx.no = b.no + 1
			bx.timeSignatureChange = false
			bx.position = b.End()
			//s.bars = append(s.bars, bx)
			s.addBar(bx)
			b = bx
		}

		bnew.no = b.no + 1
		s.addBar(bnew)
		//s.bars = append(s.bars, bnew)
		b = bnew
	}

	//	fmt.Printf("s.endPos: %v b.End(): %v\n", s.endPos, b.End())

	var diffTicks uint
	var missingBars uint

	if b.End() < s.endPos {
		diffTicks = uint(s.endPos - b.End())
		missingBars = diffTicks / b.Length()
	}

	for i := 0; i < int(missingBars); i++ {
		bx := b.dup()
		bx.no = b.no + 1
		bx.timeSignatureChange = false
		bx.position = b.End()
		//s.bars = append(s.bars, bx)
		s.addBar(bx)
		b = bx
	}

	return nil
}

func (s *Score) scanMeterAndKeyChanges() error {
	// s.currentTrackEndPos
	evts, err := s.smf.MeterTrack()

	if err != nil {
		return err
	}

	for _, ev := range evts {
		var text string
		var num, denom uint8
		var key midismf.Key

		switch {
		case ev.Message.GetMetaKey(&key):
			s.keyChanges[ev.Position.Uint()] = key
		//	fmt.Printf("got key change: %s\n", key.String())
		case ev.Message.GetMetaMarker(&text):
			if !strings.HasPrefix(text, "//") {
				//	fmt.Printf("got marker: %q\n", text)
				s.markers[ev.Position.Uint()] = text
			}
			//		case ev.Message.GetMetaCuepoint(&text):
			//		if !strings.HasPrefix(text, "//") {
			//			fmt.Printf("got cuepoint: %q\n", text)
			//			s.markers[ev.Position.Uint()] = text
			//		}
		case ev.Message.GetMetaMeter(&num, &denom):
			//	fmt.Printf("got meter change: %v/%v\n", num, denom)
			if ev.Position == 0 {
				s.startingTimeSignature = [2]uint8{num, denom}
			}
			s.timeSignatureChanges = append(s.timeSignatureChanges, [3]uint{ev.Position.Uint(), uint(num), uint(denom)})
		}

	}

	return nil
}

func (s *Score) scanTempoChanges() error {
	// s.currentTrackEndPos
	evts, err := s.smf.TempoTrack()

	if err != nil {
		return err
	}

	for _, ev := range evts {
		var bpm float64
		if ev.Message.GetMetaTempo(&bpm) {
			if ev.Position == 0 {
				s.startingTempo = bpm
			}
			s.tempoChanges = append(s.tempoChanges, map[smf.TicksAbsPos]float64{ev.Position: bpm})
		}
	}

	_ = evts
	return nil
}

func (s *Score) WritePDF(file path.Local) error {
	if path.IsDir(file) {
		return fs.ErrExpectedFile.Params(file.String())
	}

	bk, err := s.makeBook()
	if err != nil {
		return err
	}
	return bk.ToPDF(file, items.DEBUG)
}

func (s *Score) String() (string, error) {
	bk, err := s.makeBook()

	if err != nil {
		return "", err
	}

	return bk.String(), nil
}

func (s *Score) makeBook() (*lilypond.Book, error) {

	s.ticksPerQN = s.smf.TicksPerQN()

	//	fmt.Println("running through the tracks")

	var tracks track.Tracks

	for trackName := range s.score.Unrolled {
		tr, err := s.score.GetTrack(trackName)
		if err != nil {
			return nil, fmt.Errorf("can't get the track %q", trackName)
		}
		tracks = append(tracks, tr)
	}

	sort.Sort(tracks)

	var smfTracks []midismf.Track

	for _, track := range tracks {

		if s.fileGroup != "*" && track.FileGroup != s.fileGroup {
			continue
		}

		// ignore tracks without a MIDIchannel
		if track.MIDIChannel < 0 {
			continue
		}

		s.currentBarNo = 0
		s.currentBarTick = 0
		s.currentTempo = 0
		s.currentTimeSignature = s.startingTimeSignature
		s.lastEvTick = 0

		smfTrack, endpos, err := s.smf.ConvertTrack(track)

		if endpos > s.endPos {
			s.endPos = endpos
		}

		if err != nil {
			return nil, fmt.Errorf("could not convert score track to smf track as part of lilypond conversion")
		}

		smfTracks = append(smfTracks, smfTrack)
		tr := s.newTrack(track.Name, track.MIDIChannel, track.LilyPondClef, track.LilyPondDynamics)
		s.tracks = append(s.tracks, tr)
	}

	/*
	   we needed to run through the tracks before, in order to get the s.endPos
	*/

	//	fmt.Println("scanMeterChanges")

	// s.currentTrackEndPos = endpos

	err := s.scanMeterAndKeyChanges()
	if err != nil {
		return nil, err
	}

	//	fmt.Println("scanTempoChanges")

	err = s.scanTempoChanges()
	if err != nil {
		return nil, err
	}

	//	fmt.Println("setBars")

	err = s.setBars()

	if err != nil {
		return nil, err
	}

	s.setMarkesAndKeyChangesToBars()

	//	fmt.Println("now adding events")

	for i, smftrack := range smfTracks {
		s.lastEvTick = 0
		s.lastOpenedNote = 0

		tr := s.tracks[i]

		for _, ev := range smftrack {
			s.addEventToTrack(tr, ev)
		}

		if tr.isPercussion() && s.lastOpenedNote != 0 {
			s.closeNote(tr, s.lastOpenedNote)
			s.lastOpenedNote = 0
		}

	}

	g := new(lilypond.StaffGroup).Loose()

	for i, tr := range s.tracks {
		tr.registerAllMyNotes(s, i == 0)
		g.Add(tr.staff)
	}

	bk := lilypond.NewBook(g)

	bk.Paper.Add(lilypond.Line(`
      #(set-paper-size "a4")
  		system-system-spacing =
   			#'((basic-distance . 16)
     			(minimum-distance . 8)
    			(padding . 1.5)
    			(stretchability . 60)) % defaults: 16, 8, 1.8, 60
	`))

	bk.Score.Layout.Add(lilypond.Line(` 
		#(layout-set-staff-size 12)
    %% \magnifyStaff #(magstep -4)
    
    \context {
      \Staff
      %%% Für 2.19.80:
      %% \magnifyStaff #(magstep -4)
      %% fontSize = -4
      %% \override StaffSymbol.staff-space = #(magstep -4)
      %% \override StaffSymbol.thickness = #(magstep -4)
      %% \override VerticalAxisGroup.default-staff-staff-spacing.basic-distance = #4
    }
		`))

	/*
			 \context {
		    \Staff
		    %%% Für 2.19.80:
		    % \magnifyStaff #(magstep -3)
		    %%% Für 2.18.2:
		    fontSize = -3
		    \override StaffSymbol.staff-space = #(magstep -3)
		    \override StaffSymbol.thickness = #(magstep -3)
		  }
	*/

	if s.composer != "" {
		bk.Header.SetComposer(s.composer)
	}

	if s.title != "" {
		bk.Header.SetTitle(s.title)
	}

	if s.copyright != "" {
		bk.Header.SetCopyright(s.copyright)
		bk.Header.SetTagline("")
	} else {
		bk.Header.SetTaglineOff()
	}

	if s.poet != "" {
		bk.Header.SetPoet(s.poet)
	}

	if s.arranger != "" {
		bk.Header.SetArranger(s.arranger)
	}

	if s.subtitle != "" {
		bk.Header.SetSubTitle(s.subtitle)
	}

	return bk, nil
}

/*
// the example from here:
	// https://lilypond.org/doc/v2.23/Documentation/learning/four_002dpart-satb-vocal-score
	// transscribed without relative
	var (
		global   = ElementGroup(KeyMajor(D(0)), TimeSignature{4, 4})
		sopranos = new(Staff).SetName("sopranos").With(InstrumentName("Soprano"))
		altos    = new(Staff).SetName("altos").With(InstrumentName("Alto"))
		tenors   = new(Staff).SetName("tenors").With(InstrumentName("Tenor"))
		basses   = new(Staff).SetName("basses").With(InstrumentName("Bass"))
		lh       = new(Staff).SetName("lh")
		rh       = new(Staff).SetName("rh")

		ch = new(ChoirStaff).Add(sopranos, altos, tenors, basses)
		ps = new(PianoStaff).Set(lh, rh).With(InstrumentName("Piano"))
		g  = new(StaffGroup).Add(ch, ps).Loose() // Loose prevents the "StaffGroup" string from being printed
	)

	words := ElementGroup(
		String("Wor -- thy"), BarChange,
		String("is the lamb"), BarChange,
		String("that was slain"), BarChange,
	)

	sopranos.NewVoice().SetName("sopranos").Add(ClefTreble(0), global,
		Rest().Quarter(), D(2).Half(), A(1).Quarter(), BarChange,
		D(2).QuarterDot(1), D(2).Eightth(), A(1).Half(), BarChange,
		C(2).Is().Quarter(), D(2), C(2).Is().Half(), BarChange,
	).NewLyrics().Add(words)

	altos.NewVoice().SetName("altos").Add(ClefTreble(0), global,
		Rest().Quarter(), A(1).Half(), A(1).Quarter(), BarChange,
		F(1).Is().QuarterDot(1), F(1).Is().Eightth(), A(1).Half(), BarChange,
		G(1).Quarter(), F(1).Is(), F(1).Is().Half(), BarChange,
	).NewLyrics().Add(words)

	tenors.NewVoice().SetName("tenors").Add(Clef_G(8), global,
		Rest().Quarter(), F(1).Is().Half(), E(1).Quarter(), BarChange,
		D(1).QuarterDot(1), D(1).Eightth(), D(1).Half(), BarChange,
		E(1).Quarter(), A(0), C(1).Is().Half(), BarChange,
	).NewLyrics().Add(words)

	basses.NewVoice().SetName("basses").Add(ClefBass(0), global,
		Rest().Quarter(), D(1).Half(), C(1).Is().Quarter(), BarChange,
		B(0).QuarterDot(1), B(0).Eightth(), F(0).Is().Half(), BarChange,
		E(0).Quarter(), D(0), A(0).Half(), BarChange,
	).NewLyrics().Add(words)

	lh.NewVoice().Add(ClefTreble(0), global,
		Rest().Quarter(),
		new(Chord).Add(A(1), D(2), F(2).Is()).SetDuration(Half()),
		new(Chord).Add(A(1), E(2), A(2)).SetDuration(Quarter()),
		BarChange,
		new(Chord).Add(D(3), F(2).Is(), D(2)).SetDuration(QuarterDot(1)),
		new(Chord).Add(D(3), F(2).Is(), D(2)).SetDuration(Eightth()),
		new(Chord).Add(A(1), D(2), A(2)).SetDuration(Half()),
		BarChange,
		new(Chord).Add(G(1), C(2).Is(), G(2)).SetDuration(Quarter()),
		new(Chord).Add(A(1), D(2), F(2).Is()).SetDuration(Quarter()),
		new(Chord).Add(A(1), C(2).Is(), E(2)).SetDuration(Half()),
		BarChange,
	)

	rh.NewVoice().Add(ClefBass(0), global,
		new(Chord).Add(D(0), D(-1)).SetDuration(Quarter()),
		new(Chord).Add(D(0), D(-1)).SetDuration(Half()),
		new(Chord).Add(C(0).Is(), C(-1).Is()).SetDuration(Quarter()),
		BarChange,
		new(Chord).Add(B(0), B(-1)).SetDuration(QuarterDot(1)),
		new(Chord).Add(B(-2), B(-1)).SetDuration(Eightth()),
		new(Chord).Add(F(0).Is(), F(-1).Is()).SetDuration(Half()),
		BarChange,
		new(Chord).Add(E(0), E(-1)).SetDuration(Quarter()),
		new(Chord).Add(D(0), D(-1)).SetDuration(Quarter()),
		new(Chord).Add(A(0), A(-1)).SetDuration(Half()),
		BarChange,
	)

	bk := NewBook(g)

	//	fmt.Println(bk.String())

	err := bk.ToPDF(path.MustWD().Join("TestChoir"))

	if err != nil {
		t.Errorf("error while producing pdf: %s", err.Error())
	}
*/
