package score

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gitlab.com/gomidi/muskel/file"
	"gitlab.com/gomidi/muskel/filter"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/sketch"
	"gitlab.com/gomidi/muskel/table"
	"gitlab.com/gomidi/muskel/timbre"
	"gitlab.com/gomidi/muskel/track"
	"gitlab.com/gomidi/muskel/tuning"
)

type Option func(s *Score)

func Column(colname string) Option {
	return func(s *Score) {
		s.mainCol = colname
	}
}

func Debug() Option {
	return func(s *Score) {
		//s.mainCol = colname
		sketch.DEBUG = true
	}
}

func Sketch(sketchname string) Option {
	return func(s *Score) {
		s.mainSketch = sketchname
	}
}

func PrintBarComments() Option {
	return func(s *Score) {
		s.printBarComments = true
	}
}

type Score struct {
	mainFile       string
	mainSketch     string
	mainCol        string
	params         []string // params must have the syntax [trackname]#[no]:[value] where no is the params number, e.g. voc#2:c#'
	includedScores map[string]*Score
	lyrics         map[string][]string
	properties     map[string]interface{}
	tokens         map[string]string
	IsUnrolled     bool
	exclSketch     map[string]string

	printBarComments bool

	Bars     []*sketch.Bar
	Parts    map[string][2]uint
	Tracks   map[string]*track.Track
	Tunings  map[string]*tuning.Tuning
	Timbres  map[string]*timbre.Timbre
	Scales   map[string]items.Mode
	Filters  map[string]*filter.Filter
	Sketches map[string]*sketch.Sketch
	Unrolled map[string][]*items.Event
	Files    map[string]*file.File
	Parent   *Score
}

func (s *Score) FilterTrack(colName string, events []*items.Event) []*items.Event {
	return events
}

func New(filepath string, params []string, options ...Option) *Score {
	s := &Score{
		Tracks:         map[string]*track.Track{},
		Tunings:        map[string]*tuning.Tuning{},
		Timbres:        map[string]*timbre.Timbre{},
		Scales:         map[string]items.Mode{},
		Filters:        map[string]*filter.Filter{},
		Sketches:       map[string]*sketch.Sketch{},
		Files:          map[string]*file.File{},
		properties:     map[string]interface{}{},
		Unrolled:       map[string][]*items.Event{},
		Parts:          map[string][2]uint{},
		tokens:         map[string]string{},
		includedScores: map[string]*Score{},
		lyrics:         map[string][]string{},
		mainFile:       filepath,
		mainSketch:     "=SCORE",
		params:         params,
		exclSketch:     map[string]string{},
	}

	for _, opt := range options {
		opt(s)
	}
	return s
}

func (sc *Score) Properties() map[string]interface{} {
	return sc.properties
}

func (sc *Score) AddLyrics(l map[string][]string) {
	for k, v := range l {
		sc.lyrics[k] = v
	}
}

func (sc *Score) Lyric(part string, fromLine, toLine int) (tokens []string, err error) {
	//fmt.Printf("Lyrics %q[%v:%v] // %#v", part, fromLine, toLine, sc.lyrics)
	p, has := sc.lyrics[part]

	if !has {
		return nil, fmt.Errorf("could not find lyrics for %q", part)
	}
	//	l := strings.Split(p, "\n")

	if fromLine < 0 {
		fromLine = 0
	}

	if fromLine >= len(p) {
		return
	}

	if toLine > len(p) || toLine <= 0 {
		toLine = len(p)
	}

	for i := fromLine; i < toLine; i++ {
		d := p[i]
		d = strings.Replace(d, "-", "- ", -1)
		d = strings.Replace(d, "_", "_ ", -1)
		d = strings.Replace(d, "\t", " ", -1)
		d = strings.Replace(d, "\n", " ", -1)
		ds := strings.Split(d, " ")

		for _, dd := range ds {
			if dd != "" {
				tokens = append(tokens, dd)
			}
		}
	}

	return
}

func (sc *Score) AddInclude(filepath string, sketch string, params []string) error {
	//fmt.Printf("AddInclude %q %q\n", filepath, part)
	fname, err := sc.findInclude(filepath)
	if err != nil {
		return fmt.Errorf("can't find include %q", filepath)
	}

	sco, has := sc.includedScores[fname]
	if !has {
		err := sc.Include(fname, sketch, params)
		if err != nil {
			return fmt.Errorf("can't include %q, reason: %s", filepath, err.Error())
		}
		sco = sc.includedScores[fname]
	}
	if sketch == "" {
		for k, v := range sco.tokens {
			sc.tokens[k] = v
		}

		for skname, sk := range sco.Sketches {
			sc.Sketches[skname] = sk
		}

		sc.AddLyrics(sco.lyrics)

		for trname, tr := range sco.Tracks {
			sc.Tracks[trname] = tr
		}

		return nil
	}

	switch sketch[0] {
	case '=':
		sk, err := sco.GetSketch(sketch)
		if err != nil {
			return fmt.Errorf("can't find sketch %q in include %q, reason: %s", sketch, filepath, err.Error())
		}
		sc.Sketches[sketch] = sk
		return nil
	case '@':
		ly, has := sco.lyrics[sketch]
		if !has {
			return fmt.Errorf("can't find lyrics %q in include %q", sketch, filepath)
		}
		sc.lyrics[sketch] = ly
		return nil
	default:
		sh, err := sco.GetToken(sketch)
		if err != nil {
			return fmt.Errorf("can't find token %q in include %q", sketch, filepath)
		}
		sc.tokens[sketch] = sh
		return nil
	}

}

func (sc *Score) AddToken(key string, value string) {
	if reservedNames[key] {
		panic(fmt.Sprintf("can't add token %q: reserved name", key))
	}
	//fmt.Printf("adding token: %q -> %q\n", key, value)
	sc.tokens[key] = value
}

func (sc *Score) GetToken(name string) (string, error) {
	//fmt.Printf("find token: %q in %v\n", name, sc.tokens)
	tk, has := sc.tokens[name]
	if !has {
		if sc.Parent != nil {
			return sc.Parent.GetToken(name)
		}
		return "", fmt.Errorf("can't find token %q", name)
	}
	return tk, nil
}

func (sc *Score) AddProperty(key, value string) {
	//fmt.Printf("adding property: %q -> %q\n", key, value)
	sc.properties[key] = value
}

func (sc *Score) HasTrack(track string) bool {
	_, has := sc.Tracks[track]
	return has
}

func (sc *Score) GetTrack(track string) (*track.Track, error) {
	track = strings.Trim(track, "!")
	// fmt.Printf("looking for track %q\n", track)
	tr, has := sc.Tracks[track]
	if !has {
		// panic(fmt.Sprintf("can't find track %q", track))
		return nil, fmt.Errorf("can't find track %q", track)
	}
	return tr, nil
}

func (sc *Score) parse(fname string, sco *Score) error {
	f, err := file.New(fname, sco)
	if err != nil {
		return err
	}
	sc.Files[fname] = f
	return f.Parse()
}

func (sc *Score) Parse() error {
	return sc.parse(sc.mainFile, sc)
}

func (sc *Score) Delay(trackName string) (del [2]int) {
	return sc.Tracks[trackName].Delay
}

func (sc *Score) Channel(trackName string) (ch int8) {
	return sc.Tracks[trackName].MIDIChannel
}

func (sc *Score) Program(trackName string) (prog int8) {
	return sc.Tracks[trackName].MIDIProgram
}

func (sc *Score) Volume(trackName string) (vol int8) {
	return sc.Tracks[trackName].MIDIVolume
}
func (sc *Score) Bank(trackName string) (bank int8) {
	return sc.Tracks[trackName].MIDIBank
}
func (sc *Score) Transpose(trackName string) (track int8) {
	return sc.Tracks[trackName].MIDITranspose
}
func (sc *Score) PitchBendRange(trackName string) (pb uint8) {
	return sc.Tracks[trackName].PitchbendRange
}
func (sc *Score) VelocityScale(trackName string) (vc [5]uint8) {
	return sc.Tracks[trackName].VelocityScale
}
func (sc *Score) FileGroup(trackName string) (fg string) {
	return sc.Tracks[trackName].FileGroup
}

func (sc *Score) findInclude(filename string) (fname string, err error) {
	return findInclude(filepath.Dir(sc.mainFile), filename)
}

func (sc *Score) GetIncludedSketch(filename, sketch_table string, params []string) (*sketch.Sketch, error) {

	//fmt.Printf("GetIncludedSketch(%q,%q)\n", filename, sketch_table)
	fname, err := sc.findInclude(filename)
	if err != nil {
		return nil, fmt.Errorf("can't find include %q\n", filename)
	}

	sco, has := sc.includedScores[fname]
	if !has {
		err := sc.Include(fname, sketch_table, params)
		if err != nil {
			return nil, fmt.Errorf("can't include %q, reason: %s\n", fname, err.Error())
		}
		sco = sc.includedScores[fname]
	}
	return sco.GetSketch(sketch_table)
}

func (sc *Score) Unroll() error {
	if sc.IsUnrolled {
		return nil
	}

	var exclMode bool
	//fmt.Printf("sketchname: %q column: %q\n", sc.mainSketch, sc.mainCol)

	sketchName := sc.mainSketch
	if sketchName == "!" {
		sketchName = sc.findMostExclamedSketch()
		exclMode = true
		//fmt.Println("in exclMode")
	}

	//fmt.Printf("sketch: %q\n", sketchName)

	sketch := sc.Sketches[sketchName]
	if sketch == nil {
		return fmt.Errorf("could not find main sketch %q", sc.mainSketch)
	}

	params, err := convertParams(sc.params)
	if err != nil {
		return err
	}

	col := sc.mainCol

	if exclMode && col == "" {
		col = "!"
	}

	if col == "!" {
		col = sc.findMostExclamedCol(sketch)
	}

	switch {
	case col != "":
		_col := col
		//tr, _ := sc.GetTrack(strings.Trim(col, "!"))
		tr, _ := sc.GetTrack(col)
		if tr == nil {
			//fmt.Printf("could not find track for col %q\n", col)
			trackName := sc.findMostExclamedTrack()
			if trackName != "" {
				//fmt.Printf("taking most exclamed track %q\n", trackName)
				col = trackName
				tr = sc.Tracks[trackName]
			} else {
				//fmt.Printf("construct track %v\n", col)
				trackName = col
				tr = track.New(trackName)
				tr.MIDIChannel = 0
				sc.Tracks[trackName] = tr
			}
		}

		//fmt.Printf("col is %q\n", col)
		events, err := sketch.Unroll(tr, _col, params[_col]...)
		if err != nil {
			return fmt.Errorf("error while unrolling column %q of sketch %q with params %v: %s", col, sketch.Name, params[col], err.Error())
		}

		sc.Tracks = map[string]*track.Track{
			tr.Name: tr,
		}
		//fmt.Printf("unrolling col %q for track %q\n", col, tr.Name)
		sc.Unrolled[tr.Name] = sc.replaceScalenotesForCol(tr.Name, events)
	default:
		//fmt.Printf("%v\n", sc.Tracks)
		for c := range sketch.Columns {
			//tr, _ := sc.GetTrack(strings.Trim(c, "!"))
			tr, _ := sc.GetTrack(c)
			if tr == nil {
				//panic(fmt.Sprintf("could not find track %q\n", col))
				return fmt.Errorf("could not find track %q\n", c)
			}

			events, err := sketch.Unroll(tr, c, params[c]...)
			if err != nil {
				return fmt.Errorf("error while unrolling column %q of sketch %q with params %v: %s", c, sketch.Name, params[c], err.Error())
			}

			sc.Unrolled[tr.Name] = sc.replaceScalenotesForCol(tr.Name, events)
		}

		for trackname, tr := range sc.Tracks {
			if tr != nil && tr.Import != "" {
				events, err := sketch.Unroll(tr, tr.Import, params[trackname]...)
				if err != nil {
					return fmt.Errorf("error while unrolling track %q importing %q of sketch %q with params %v: %s", trackname, tr.Import, sketch.Name, params[trackname], err.Error())
				}

				sc.Unrolled[trackname] = sc.replaceScalenotesForCol(tr.Import, events)
			}
		}
	}

	sc.Bars, err = sketch.UnrolledBars()
	if err != nil {
		return err
	}
	sc.Parts = sketch.Parts
	sc.IsUnrolled = true

	for c, evts := range sc.Unrolled {
		sc.Unrolled[c] = sc.replaceScalenotes(c, evts)
	}

	lastBar := sc.Bars[len(sc.Bars)-1]

	endPos := lastBar.Position + uint(lastBar.Length32th())
	for _, tr := range sc.Tracks {
		tr.EndPos = endPos
	}
	return nil
}

func (sc *Score) WriteUnrolled(wr io.Writer) error {
	err := sc.Unroll()
	if err != nil {
		return err
	}

	var tracks []string
	for track := range sc.Tracks {
		tracks = append(tracks, track)
	}

	sort.Strings(tracks)

	//tr.
	//Sketch{Table: table.NewTable("=SCORE", -1, nil)}

	//fmt.Fprint(wr, "TRACKS |")
	trks := table.NewTracks(-1, nil)

	for _, tr := range tracks {
		trks.AddCol(sc.Tracks[tr].Name)
		//fmt.Fprintf(wr, " %s |", tr)
	}

	var data []string

	data = []string{"File"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].FileGroup
		data = append(data, val)
	}

	trks.AddLine(data)

	data = []string{"Channel"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].MIDIChannel
		var v string
		if val > -1 {
			v = fmt.Sprintf("%v", val+1)
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"Program"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].MIDIProgram
		var v string
		if val > -1 {
			v = fmt.Sprintf("%v", val+1)
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"Bank"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].MIDIBank
		var v string
		if val > -1 {
			v = fmt.Sprintf("%v", val+1)
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"Transpose"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].MIDITranspose
		var v string
		if val != 0 {
			v = fmt.Sprintf("%v", val)
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"Volume"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].MIDIVolume
		var v string
		if val > -1 {
			v = fmt.Sprintf("%v", val+1)
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"Delay"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].Delay
		var v string
		if val[0] != 0 {
			v = fmt.Sprintf("%v/%v", val[0], val[1])
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"PitchbendRange"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].PitchbendRange
		var v string
		if val > 0 {
			v = fmt.Sprintf("%v", val)
		}
		data = append(data, v)
	}

	trks.AddLine(data)

	data = []string{"VelocityScale"}
	for _, tr := range tracks {
		val := sc.Tracks[tr].VelocityScale
		var v string
		// "min: %v max: %v random: %v step: %v center: %v"
		//0: min 1: max 2: randomize-factor 3: step-width, 4: center
		v = fmt.Sprintf("min: %v max: %v random: %v step: %v center: %v", val[0], val[1], val[2], val[3], val[4])
		data = append(data, v)
	}
	trks.AddLine(data)

	data = []string{"Import"}
	for _, tr := range tracks {
		//fmt.Printf("track %q import: %q\n", tr, sc.Tracks[tr].Import)
		val := sc.Tracks[tr].Import
		data = append(data, val)
	}

	trks.AddLine(data)

	fm := &formatter{wr}
	trks.WriteTo(fm)

	fmt.Fprintf(wr, "\n")
	//fmt.Fprintf(wr, "\n\n")

	sk := &table.Sketch{Table: table.NewTable("=SCORE", -1, nil)}
	// sk := table.NewSketch("score", -1, nil)
	_ = sk

	var cols []string

	for col := range sc.Unrolled {
		cols = append(cols, col)
	}

	sort.Strings(cols)

	var score = map[uint][][]string{}

	for colNo, col := range cols {
		sk.AddCol(col)

		for _, ev := range sc.Unrolled[col] {
			//fmt.Printf("[%v] ev: %v\n", ev.Position, ev)
			m := score[ev.Position]
			//fmt.Printf("len(score[%v]) = %v // colNo: %v\n", ev.Position, len(m), colNo)
			if len(m) == 0 || colNo == 0 || m[len(m)-1][colNo] != "" {
				//if colNo == 0 {
				//m = make([][]string, len(cols))
				//m = [][]string{}
				cm := make([]string, len(cols))
				//fmt.Printf("appending %#v to %#v\n", cm, m)
				m = append(m, cm)
			}

			//fmt.Printf("setting score[%v][%v][%v] to %q\n", ev.Position, len(m)-1, colNo, ev.String())
			m[len(m)-1][colNo] = ev.String()
			/*
				if cm, hasCol := m[colNo]; hasCol {

				}
			*/

			//m[colNo] = append(m[colNo], ev.String())
			score[ev.Position] = m
		}
	}

	var sorted sortedLines

	//fmt.Printf("score: %#v\n", score)

	for pos, cls := range score {
		for _, _cls := range cls {
			//fmt.Printf("adding line %v\n", line{pos: pos, cols: _cls})
			sorted = append(sorted, line{pos: pos, cols: _cls})
		}
	}

	lastTimeSig := [2]uint8{4, 4}
	var lastTempoChange float64 = 0
	var lastScale = &items.Scale{BaseNote: 60, Mode: items.Major}

	//fmt.Printf("sorted: %#v\n", sorted)

	for _, bar := range sc.Bars {
		if bar.Include != nil {
			return fmt.Errorf("bar with include not allowed here bar no %v ", bar.No)
		}
		//fmt.Printf("// #%v\n", bar.No)
		//fmt.Printf("[%v] bar %s\n", bar.No, bar.String())
		s := fmt.Sprintf("#%s", bar.Part)
		if bar.TimeSig != lastTimeSig || bar.No == 0 {
			s += fmt.Sprintf(" %v/%v", bar.TimeSig[0], bar.TimeSig[1])
			lastTimeSig = bar.TimeSig
		}
		if (bar.TempoChange > 0 && bar.TempoChange != lastTempoChange) || bar.No == 0 {
			if bar.TempoChange > 0 {
				s += fmt.Sprintf(" @%0.2f", bar.TempoChange)
				lastTempoChange = bar.TempoChange
			} else {
				var tempoChange float64 = 120
				s += fmt.Sprintf(" @%0.2f", tempoChange)
				lastTempoChange = tempoChange
			}
		}

		if ((bar.Scale) != nil && bar.Scale != lastScale) || bar.No == 0 {
			if bar.Scale != nil {
				s += fmt.Sprintf(" %s", bar.Scale.String())
				lastScale = bar.Scale
			} else {
				s += fmt.Sprintf(" %s", lastScale.String())
			}
		}

		if sc.printBarComments {
			var comment = bar.Comment
			comment += fmt.Sprintf(" #%v", bar.No+1)
			s += fmt.Sprintf("           // %s", comment)
		}

		sorted = append(sorted, line{pos: bar.Position, cols: []string{s}})

		for relPos, tc := range bar.InnerTempoChanges {
			_ = tc
			lastTempoChange = tc
			sorted = append(sorted, line{pos: bar.Position + relPos, cols: []string{}})
		}

		for relPos, sc := range bar.InnerScales {
			_ = sc
			lastScale = sc
			sorted = append(sorted, line{pos: bar.Position + relPos, cols: []string{}})
		}
	}

	//fmt.Printf("sorted: %#v\n", sorted)

	sort.Sort(sorted)
	//sk.AddLine([]string{"#"})

	var lastBarStart uint

	for _, lin := range sorted {
		if len(lin.cols) == 1 && len(lin.cols[0]) > 0 && lin.cols[0][0] == '#' {
			sk.AddLine([]string{lin.cols[0]})
			lastBarStart = lin.pos
		} else {
			relPos := lin.pos - lastBarStart
			l := []string{items.Pos32thToString(lin.pos - lastBarStart)}
			bidx := sc.GetBarIdxOf(lastBarStart)
			if bidx >= 0 && len(lin.cols) == 0 {
				if tc, hasTc := sc.Bars[bidx].InnerTempoChanges[relPos]; hasTc {
					l[0] = l[0] + fmt.Sprintf(" @%0.2f", tc)
				}

				scale := sc.Bars[bidx].InnerScales[relPos]
				if scale != nil {
					l[0] = l[0] + " " + scale.String()
				}
				//l = append(l, lin.cols...)
				sk.AddLine(l)
			} else {
				l = append(l, lin.cols...)
				sk.AddLine(l)
			}
			//fmt.Printf("adding line %q\n", l)
		}
	}

	err = sk.WriteTo(fm)
	if err != nil {
		return err
	}

	fmt.Fprintf(wr, "\n")
	return nil
}

// gets the bar idx of an event at the given position
// returns -1, if no bar could be found
func (s *Score) GetBarIdxOf(pos uint) (baridx int) {
	baridx = -1

	for idx, b := range s.Bars {
		if b.Position > pos {
			break
		}
		baridx = idx
	}

	return baridx
}

func convertScaleNotes(in items.Item, scale *items.Scale) items.Item {
	if scale == nil {
		return in
	}
	switch v := in.(type) {
	case *items.Note:
		if v.IsScaleNote && scale != nil {
			nt := v.Dup().(*items.Note)
			nt.IsScaleNote = false
			nt.ScaleNote = 0
			sn := v.ScaleNote
			aug := nt.Augmenter
			key := scale.StepToNote(sn)
			if aug == "#" {
				key++
			}
			nt.Letter, nt.Augmenter, nt.Octave = items.KeyToNote(key)
			//fmt.Printf("convert scalenote %v to %s via %#v\n", v.ScaleNote-1, nt.String(), scale)
			return nt
		} else {
			return v
		}
	case *items.NTuple:
		n := v.Dup().(*items.NTuple)
		for k, it := range n.Events {
			it.Item = convertScaleNotes(it.Item, scale)
			n.Events[k] = it
		}
		return n

	case *items.MultiItem:
		mv := v.Dup().(*items.MultiItem)

		for k, it := range mv.Events {
			_it := it
			_it.Item = convertScaleNotes(it.Item, scale)
			mv.Events[k] = _it
		}
		return mv
	default:
		return v
	}
}

func (s *Score) replaceScalenotesForCol(colname string, evts []*items.Event) []*items.Event {
	var res []*items.Event

	//fmt.Printf("replaceScalenotesForCol %q called\n", colname)

	var scale *items.Scale

	for _, ev := range evts {
		nev := ev.Dup()

		switch v := nev.Item.(type) {
		case *items.Scale:
			//fmt.Printf("got scale in col %q: %s\n", colname, v.String())
			scale = v.Dup().(*items.Scale)
			if scale.Mode == nil {
				sm, has := s.Scales[scale.Name]
				if !has {
					panic("can't find scale with name " + scale.Name)
				}
				scale.Mode = sm
			}
		default:
			if scale != nil {
				nev.Item = convertScaleNotes(nev.Item, scale)
			}
			res = append(res, nev)
		}
	}
	//panic("TODO: implement")
	return res
}

func (s *Score) GetMode(name string) items.Mode {
	return s.Scales[name]
}

func (s *Score) replaceScalenotes(colname string, evts []*items.Event) []*items.Event {
	var res []*items.Event

	scale := s.Bars[0].Scale
	if scale == nil {
		scale = &items.Scale{}
		scale.BaseNote = 60
		scale.Mode = items.Major
	}

	//fmt.Printf("replaceScalenotes %q starting with scale: %s\n", colname, scale.String())

	for _, ev := range evts {
		bidx := s.GetBarIdxOf(ev.Position)
		if s.Bars[bidx].Scale != nil {
			scale = s.Bars[bidx].Scale.Dup().(*items.Scale)
			//fmt.Printf("I scale %q\n", scale.Name)
			//fmt.Printf("replaceScalenotes %q found scale: %s at bar %s\n", colname, scale.String(), s.Bars[bidx].String())
			if scale.Mode == nil {
				sm, has := s.Scales[scale.Name]
				if !has {
					panic("can't find scale with name " + scale.Name)
				}
				scale.Mode = sm
			}
		}

		barPos := s.Bars[bidx].Position

		if len(s.Bars[bidx].InnerScales) != 0 {
			var scChanges internalScaleChanges

			for pos, sc := range s.Bars[bidx].InnerScales {
				scChanges = append(scChanges, internalScaleChange{relPos: pos, scale: sc})
			}

			sort.Sort(scChanges)

			for _, scCh := range scChanges {
				if scCh.relPos+barPos > ev.Position {
					break
				}

				scale = scCh.scale.Dup().(*items.Scale)
				//fmt.Printf("II scale %q\n", scale.Name)
				if scale.Mode == nil {
					sm, has := s.Scales[scale.Name]
					if !has {
						panic("can't find scale with name " + scale.Name)
					}
					scale.Mode = sm
				}
			}
		}

		//fmt.Printf("event: %s scale: %v\n", ev, scale)

		nev := ev.Dup()
		nev.Item = convertScaleNotes(nev.Item, scale)
		res = append(res, nev)
	}
	//panic("TODO: implement")
	return res
}

func (sc *Score) AddTrack(t *track.Track) {
	// fmt.Printf("AddTrack %q\n", t.Name)
	sc.Tracks[t.Name] = t
}
func (sc *Score) AddTuning(t *tuning.Tuning) {
	sc.Tunings[t.Name] = t
}

func (sc *Score) AddFilter(t *filter.Filter) {
	sc.Filters[t.Name] = t
}

func (sc *Score) AddScale(t items.Mode) {
	//fmt.Printf("AddScale: %s\n", t.Name())
	if t.Name() == "minor" {
		panic("minor added")
	}
	sc.Scales[t.Name()] = t
}

func (sc *Score) AddTimbre(t *timbre.Timbre) {
	sc.Timbres[t.Name] = t
}

func (sc *Score) AddSketch(name string) interface {
	ParseLine([]string) error
	AddColumn(string)
} {
	//fmt.Printf("AddSketch %q called\n", name)
	if idx := strings.Index(name, "!"); idx > 0 {
		excl := name[idx:]
		name = name[:idx]
		sc.exclSketch[excl] = name
	}
	sk := sketch.NewSketch(name, sc)
	sk.File = sc.mainFile
	sc.Sketches[name] = sk
	return sk
}

func (sc *Score) findMostExclamedSketch() (name string) {
	var best int
	for excl, sketch := range sc.exclSketch {
		if len(excl) > best {
			name = sketch
			best = len(excl)
		}
	}
	return
}

func (sc *Score) findMostExclamedCol(sketch *sketch.Sketch) (col string) {
	var best int = -1
	for c := range sketch.Columns {
		if idx := strings.Index(c, "!"); idx > 0 {
			if len(c[idx:]) > best {
				col = c
			}
		}
	}

	if best == -1 {
		col = sketch.FirstColumn()
	}

	return
}

func (sc *Score) findMostExclamedTrack() (track string) {
	var best int = -1
	for tr := range sc.Tracks {
		if best == -1 {
			track = tr
			best = 0
		}
		if idx := strings.Index(tr, "!"); idx > 0 {
			if len(tr[idx:]) > best {
				track = tr
			}
		}
	}

	return
}

func (sc *Score) GetSketch(name string) (*sketch.Sketch, error) {
	/*
		if name[0] != '=' {
			panic("invalid sketch name " + name)
		}
	*/
	s := sc.Sketches[name]

	if s == nil {
		return nil, fmt.Errorf("can't find sketch %q\n", name)
	}

	s.File = sc.mainFile

	return s, nil
}

func mkRegexp(s string) (*regexp.Regexp, error) {
	s = strings.ReplaceAll(s, ".", "\\.")
	s = strings.ReplaceAll(s, "*", ".*")
	s = "^" + s + "$"
	return regexp.Compile(s)
}

func getTokenShorty(token string) string {
	if idx := strings.Index(token[1:], "."); idx > 0 && len(token) > idx+2 {
		return token[idx+1:]
	}
	return token
}

func (sc *Score) embed(patt string) error {
	//fmt.Printf("embed: %q\n", patt)
	if idx := strings.Index(patt, "*"); idx >= 0 {
		reg, err := mkRegexp(patt)
		if err != nil {
			return fmt.Errorf("invalid pattern syntax for embedding: %q", patt)
		}
		for k, v := range sc.tokens {
			if reg.MatchString(k) {
				//fmt.Printf("register I: %q\n", getTokenShorty(k))
				sc.tokens[getTokenShorty(k)] = v
			}
		}
		return nil
	}

	for k, v := range sc.tokens {
		if k == patt {
			//fmt.Printf("register II: %q\n", getTokenShorty(k))
			sc.tokens[getTokenShorty(k)] = v
		}
	}

	return nil
}

func (sc *Score) Embed(patterns ...string) error {
	for _, patt := range patterns {
		err := sc.embed(patt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Score) Include(filename string, sketch string, params []string) error {
	fname, err := sc.findInclude(filename)
	if err != nil {
		return err
	}
	sco, has := sc.includedScores[fname]
	if !has {
		var opts []Option
		if sketch != "" {
			opts = append(opts, Sketch(sketch))
		}
		sco = New(filename, params, opts...)
		sco.Parent = sc
		err := sc.parse(fname, sco)
		if err != nil {
			return err
		}
		sc.includedScores[fname] = sco
		//fmt.Printf("including %q\n", fname)
	}

	return nil
}
func (sc *Score) Format() error {
	for _, fl := range sc.Files {
		err := fl.Format()
		if err != nil {
			return err
		}
	}
	return nil
}
