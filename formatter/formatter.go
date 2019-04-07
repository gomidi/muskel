package formatter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

type Formatter struct {
	score               *score.Score
	writerSysLine       int
	jumpInLineBefore    bool
	HideTrackProperties bool
	HideHeader          bool
	bf                  strings.Builder
}

func New(s *score.Score) *Formatter {
	return &Formatter{score: s}
}

func (p *Formatter) printProp(instr *score.Track, prop interface{}) string {
	ipad := instr.Pad(fmt.Sprintf("%v", prop))
	if p.score.SmallColumns {
		return fmt.Sprintf("%v|", ipad)
	}

	return fmt.Sprintf(" %v |", ipad)
}

func (p *Formatter) writeTrackProps(bf io.Writer) {
	l := "File     |"
	if p.score.SmallColumns {
		l = "File    |"
	}

	for _, instr := range p.score.Tracks {
		if p.score.SmallColumns {
			l += fmt.Sprintf("%v|", instr.Pad(instr.FileGroup))
		} else {

			l += fmt.Sprintf(" %v |", instr.Pad(instr.FileGroup))
		}
	}

	p.writeBodyLine(bf, l)

	l = "Ch       |"
	if p.score.SmallColumns {
		l = "Ch      |"
	}

	for _, instr := range p.score.Tracks {
		if instr.MIDIChannel == -1 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, instr.MIDIChannel+1)
		}
	}

	p.writeBodyLine(bf, l)

	l = "Bank     |"
	if p.score.SmallColumns {
		l = "Bank    |"
	}

	for _, instr := range p.score.Tracks {
		if instr.MIDIBank == -1 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, instr.MIDIBank+1)
		}
	}

	p.writeBodyLine(bf, l)

	l = "Prog     |"
	if p.score.SmallColumns {
		l = "Prog    |"
	}

	for _, instr := range p.score.Tracks {
		if instr.MIDIProgram == -1 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, instr.MIDIProgram+1)
		}
	}

	p.writeBodyLine(bf, l)

	l = "Vol      |"
	if p.score.SmallColumns {
		l = "Vol     |"
	}

	for _, instr := range p.score.Tracks {
		if instr.MIDIVolume == -1 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, instr.MIDIVolume+1)
		}
	}

	p.writeBodyLine(bf, l)

	l = "PbRange  |"
	if p.score.SmallColumns {
		l = "PbRange |"
	}

	for _, instr := range p.score.Tracks {
		if instr.PitchbendRange == 0 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, instr.PitchbendRange)
		}
	}

	p.writeBodyLine(bf, l)

	l = "Trans    |"
	if p.score.SmallColumns {
		l = "Trans   |"
	}

	for _, instr := range p.score.Tracks {
		if instr.MIDITranspose == 0 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, instr.MIDITranspose)
		}
	}

	p.writeBodyLine(bf, l)

	l = "Delay    |"
	if p.score.SmallColumns {
		l = "Delay   |"
	}

	for _, instr := range p.score.Tracks {
		if instr.Delay[0] == 0 {
			l += p.printProp(instr, " ")
		} else {
			l += p.printProp(instr, fmt.Sprintf("%v/%v", instr.Delay[0], instr.Delay[1]))
		}
	}

	p.writeBodyLine(bf, l)

}

func (p *Formatter) writeTrackLines(bf io.Writer) {
	// 9 whitespace to first pipe
	l := "         |"
	if p.score.SmallColumns {
		l = "        |"
	}

	for _, instr := range p.score.Tracks {
		instr.CalcColWidth(p.score.IsUnrolled)
		//instrColWidths[i] = instr.colWidth
		if p.score.SmallColumns {
			l += fmt.Sprintf("%s|", instr.Pad("<"+instr.Name+">"))
		} else {

			l += fmt.Sprintf(" %s |", instr.Pad("<"+instr.Name+">"))
		}
	}

	p.writeBodyLine(bf, l)

	if !p.HideTrackProperties {
		p.writeTrackProps(bf)
	}
}

func (f *Formatter) printSorted(bf io.Writer, format string, m map[string]string) {
	var keys = make([]string, len(m))

	var i int
	for k := range m {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(bf, format, k, m[k])
	}
}

func (f *Formatter) prepareBars() {
	if f.score.IsUnrolled {
		for _, b := range f.score.Bars {
			b.Events = map[score.BarEventKey]*score.Event{}

			for i, instr := range f.score.Tracks {
				for _, ev := range instr.Unrolled {
					if ev.BarNo == b.BarNo {
						b.AddInstrEvent(ev.DistanceToStartOfBarIn32th, i, ev)
					}
				}
			}
		}
		return
	}

	// !f.score.isUnrolled
	for _, b := range f.score.Bars {
		b.Events = map[score.BarEventKey]*score.Event{}

		for i, instr := range f.score.Tracks {
			for j, bev := range instr.Events {
				if j == b.BarNo {
					for bidx, ev := range bev {
						b.AddInstrEvent(b.Positions[bidx], i, ev)
					}
				}
			}
		}
	}
}

func (p *Formatter) printBody(bf io.Writer, barLine string) {
	var l string
	p.jumpInLineBefore = true
	p.writeTrackLines(bf)
	//	fmt.Printf("num bars: %v\n", len(p.score.Bars))
	//	for b, bar := range p.score.Bars {
	//		fmt.Printf("bar %v: %#v\n", b, bar)
	//	}
	p.prepareBars()

	var prefix string

	if !p.score.SmallColumns {
		prefix = " "
	}

	emptyBarCounter := 0

	//fmt.Printf("len(bars) = %v\n", len(p.score.Bars))
	// fmt.Printf("parts = %#v\n", p.score.Parts)

	for _, bar := range p.score.Bars {

		if bar.IsEmpty {
			emptyBarCounter++
			continue
		}

		if emptyBarCounter > 0 {
			l = fmt.Sprintf("*%v", emptyBarCounter+1)
			p.writeBodyLine(bf, l)
			p.jumpInLineBefore = true
			emptyBarCounter = 0
		}

		if bar.JumpTo != "" {
			l = fmt.Sprintf("[%s]", bar.JumpTo)
			p.writeBodyLine(bf, l)
			p.jumpInLineBefore = true
			continue
		}

		if bar.Include != "" {
			l = fmt.Sprintf("$include(%q)", bar.Include)
			p.writeBodyLine(bf, l)
			p.jumpInLineBefore = true
			continue
		}

		if bar.TimeSigChange[0] != 0 {
			l = fmt.Sprintf("%v/%v", bar.TimeSigChange[0], bar.TimeSigChange[1])
		} else {
			l = ""
		}

		if bar.TempoChange > 0 {
			l += fmt.Sprintf("@%g", bar.TempoChange)
			if len(bar.Tilde) > 0 {
				l += bar.Tilde
			}
		}

		if bar.Scale != nil {
			var nt items.Note
			nt.Letter, nt.Augmenter, nt.Octave = items.KeyToNote(bar.Scale.BaseNote)
			nt.Velocity = -1
			l += fmt.Sprintf("\\%s^%v", bar.Scale.Name(), nt.String())
		}

		if !p.jumpInLineBefore || l != "" {
			p.writeBodyLine(bf, l)
		}

		p.jumpInLineBefore = false
		//			fmt.Printf("bar.originalPositions: %v\n", bar.originalPositions)

		for pi, pos := range bar.OriginalPositions {
			switch {
			case pos[0] == '&':
				l = fmt.Sprintf("     %s"+barLine, score.Pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			case strings.Index("123456789", pos[0:1]) != -1:
				l = fmt.Sprintf("    %s"+barLine, score.Pad(pos, 4))
				//l= fmt.Sprintf("    %4s+ |")
			case pos[0] == '.':
				l = fmt.Sprintf("     %s"+barLine, score.Pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			case pos[0] == ';':
				l = fmt.Sprintf("     %s"+barLine, score.Pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			}

			for instrCol, instr := range p.score.Tracks {
				if ev, has := bar.Events[score.BarEventKey{instrCol, bar.Positions[pi]}]; has && ev != nil {
					l += fmt.Sprintf(prefix+"%s"+barLine, instr.Pad(ev.OriginalData))
				} else {
					l += fmt.Sprintf(prefix+"%s"+barLine, instr.Pad(""))
				}
			}

			// fmt.Printf("bar: %v, pi: %v\n", bar.barNo, pi)

			if pi == 0 {
				for prt, br := range p.score.Parts {
					if br[0] == bar.OriginalBarNo {
						l += fmt.Sprintf(prefix+"%s", prt)
					}
				}
			}
			p.writeBodyLine(bf, l)
		}

	}

	if emptyBarCounter > 0 {
		l = fmt.Sprintf("_%v", emptyBarCounter)
		p.writeBodyLine(bf, l)
		emptyBarCounter = 0
	}

	lastSysLine := p.score.LastLine - p.score.LineWhereBodyStarts
	missingLines := lastSysLine - p.writerSysLine

	for i := 0; i < missingLines; i++ {
		if comment, has := p.score.BodyComments[p.writerSysLine+i]; has {
			fmt.Fprint(bf, comment+"\n")
			// bf.WriteString(comment + "\n")
		}
	}

}

func (p *Formatter) printHeader(bf io.Writer) {
	metaMap := map[string]string{}

	for k, v := range p.score.Meta {
		metaMap[score.Pad("@"+k+":", 15)] = v
	}

	p.printSorted(bf, "%s %s\n", metaMap)

	fmt.Fprint(bf, "\n\n")
	// bf.WriteString("\n\n")

	for _, incl := range p.score.HeaderIncludes {
		fmt.Fprintf(bf, "$include(%q)\n", incl)
	}

	p.printSorted(bf, "/%s/%s\n", p.score.Temperament)
	fmt.Fprint(bf, "\n\n")
	// bf.WriteString("\n\n")

	templDefMap := map[string]string{}

	for k, v := range p.score.TemplateDefinitions {
		key := k

		if v.TimeSignature[1] > 0 {
			key += fmt.Sprintf("@%v/%v", v.TimeSignature[0], v.TimeSignature[1])
		}

		templDefMap[score.Pad(key+":", 15)] = v.Original
	}

	p.printSorted(bf, "%s %s\n", templDefMap)

	fmt.Fprint(bf, "\n\n")
	//bf.WriteString("\n\n")

	for _, comment := range p.score.HeaderComments {
		fmt.Fprint(bf, comment+"\n")
		// bf.WriteString(comment + "\n")
	}
}

func (p *Formatter) WriteTo(bf io.Writer) {

	if !p.HideHeader {
		p.printHeader(bf)
		fmt.Fprint(bf, "\n=\n")
		//bf.WriteString("\n=\n")
	}

	barLine := " |"
	if p.score.SmallColumns {
		barLine = "|"
	}

	p.printBody(bf, barLine)
}

func (p *Formatter) writeBodyLine(bf io.Writer, line string) {
	p._writeBodyLine(bf, line, true)
}

func (p *Formatter) _writeBodyLine(bf io.Writer, line string, lineBreak bool) {
	// we don't write comments for unrolled scores, since the lines won't match anyway
	if !p.score.IsUnrolled {
		if comment, has := p.score.BodyComments[p.writerSysLine]; has {
			p.writerSysLine++
			fmt.Fprint(bf, comment+"\n")
			// bf.WriteString(comment + "\n")
			p._writeBodyLine(bf, line, lineBreak)
			return
		}
	}
	p.writerSysLine++
	if lineBreak {
		fmt.Fprint(bf, line+"\n")
		//bf.WriteString(line + "\n")
		return
	}
	fmt.Fprint(bf, line)
	//bf.WriteString(line)
}
