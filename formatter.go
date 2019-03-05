package muskel

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type Formatter struct {
	score            *Score
	writerSysLine    int
	jumpInLineBefore bool
}

func NewFormatter(s *Score) *Formatter {
	return &Formatter{score: s}
}

func (p *Formatter) writeInstrumentLines(bf *bytes.Buffer) {
	// 9 whitespace to first pipe
	l := "         |"
	if p.score.SmallColumns {
		l = "        |"
	}

	for _, instr := range p.score.Instruments {
		instr.calcColWidth(p.score.isUnrolled)
		//instrColWidths[i] = instr.colWidth
		if p.score.SmallColumns {
			l += fmt.Sprintf("%s|", instr.pad(instr.Name))
		} else {

			l += fmt.Sprintf(" %s |", instr.pad(instr.Name))
		}
	}

	p.writeBodyLine(bf, l)

	l = "Ch       |"
	if p.score.SmallColumns {
		l = "Ch      |"
	}

	for _, instr := range p.score.Instruments {
		if instr.MIDIChannel == -1 {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad("-"))
			} else {

				l += fmt.Sprintf(" %v |", instr.pad("-"))
			}
		} else {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(fmt.Sprintf("%v", instr.MIDIChannel+1)))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIChannel+1)))
			}
		}
	}

	p.writeBodyLine(bf, l)

	l = "Bank     |"
	if p.score.SmallColumns {
		l = "Bank    |"
	}

	for _, instr := range p.score.Instruments {
		if instr.MIDIBank == -1 {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad("-"))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad("-"))
			}
		} else {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(fmt.Sprintf("%v", instr.MIDIBank+1)))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIBank+1)))
			}
		}
	}

	p.writeBodyLine(bf, l)

	l = "Prog     |"
	if p.score.SmallColumns {
		l = "Prog    |"
	}

	for _, instr := range p.score.Instruments {
		if instr.MIDIProgram == -1 {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad("-"))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad("-"))
			}
		} else {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(fmt.Sprintf("%v", instr.MIDIProgram+1)))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIProgram+1)))
			}
		}
	}

	p.writeBodyLine(bf, l)

	l = "Vol      |"
	if p.score.SmallColumns {
		l = "Vol     |"
	}

	for _, instr := range p.score.Instruments {
		if instr.MIDIVolume == -1 {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad("-"))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad("-"))
			}
		} else {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(fmt.Sprintf("%v", instr.MIDIVolume+1)))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIVolume+1)))
			}
		}
	}

	p.writeBodyLine(bf, l)

	l = "PbRange  |"
	if p.score.SmallColumns {
		l = "PbRange |"
	}

	for _, instr := range p.score.Instruments {
		if instr.PitchbendRange == 0 {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(" "))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(" "))
			}
		} else {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(fmt.Sprintf("%v", instr.PitchbendRange)))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.PitchbendRange)))
			}
		}
	}

	p.writeBodyLine(bf, l)

	l = "Trans    |"
	if p.score.SmallColumns {
		l = "Trans   |"
	}

	for _, instr := range p.score.Instruments {
		if instr.MIDITranspose == 0 {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(" "))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(" "))
			}
		} else {
			if p.score.SmallColumns {
				l += fmt.Sprintf("%v|", instr.pad(fmt.Sprintf("%v", instr.MIDITranspose)))
			} else {
				l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDITranspose)))
			}
		}
	}

	p.writeBodyLine(bf, l)
}

func (f *Formatter) printSorted(bf *bytes.Buffer, format string, m map[string]string) {
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
	if f.score.isUnrolled {
		for _, b := range f.score.Bars {
			b.events = map[barEventKey]*Event{}

			for i, instr := range f.score.Instruments {
				for _, ev := range instr.unrolled {
					if ev.BarNo == b.barNo {
						b.addInstrEvent(ev.DistanceToStartOfBarIn32th, i, ev)
					}
				}
			}
		}
		return
	}

	// !f.score.isUnrolled
	for _, b := range f.score.Bars {
		b.events = map[barEventKey]*Event{}

		for i, instr := range f.score.Instruments {
			for j, bev := range instr.events {
				if j == b.barNo {
					for bidx, ev := range bev {
						b.addInstrEvent(b.positions[bidx], i, ev)
					}
				}
			}
		}
	}
}

func (p *Formatter) printBody(bf *bytes.Buffer, barLine string) {
	var l string
	p.jumpInLineBefore = true
	p.writeInstrumentLines(bf)
	//	fmt.Printf("num bars: %v\n", len(p.score.Bars))
	//	for b, bar := range p.score.Bars {
	//		fmt.Printf("bar %v: %#v\n", b, bar)
	//	}
	p.prepareBars()

	var prefix string

	if !p.score.SmallColumns {
		prefix = " "
	}

	//fmt.Printf("len(bars) = %v\n", len(p.score.Bars))
	// fmt.Printf("parts = %#v\n", p.score.Parts)

	for _, bar := range p.score.Bars {

		if bar.jumpTo != "" {
			l = fmt.Sprintf("[%s]", bar.jumpTo)
			p.writeBodyLine(bf, l)
			p.jumpInLineBefore = true
			continue
		}

		if bar.include != "" {
			l = fmt.Sprintf("$include(%q)", bar.include)
			p.writeBodyLine(bf, l)
			p.jumpInLineBefore = true
			continue
		}

		if bar.timeSigChange[0] != 0 {
			l = fmt.Sprintf("%v/%v", bar.timeSigChange[0], bar.timeSigChange[1])
		} else {
			l = ""
		}

		if bar.tempoChange > 0 {
			l += fmt.Sprintf("@%g", bar.tempoChange)
		}

		if !p.jumpInLineBefore || l != "" {
			p.writeBodyLine(bf, l)
		}

		p.jumpInLineBefore = false
		//			fmt.Printf("bar.originalPositions: %v\n", bar.originalPositions)

		for pi, pos := range bar.originalPositions {
			switch {
			case pos[0] == '&':
				l = fmt.Sprintf("     %s"+barLine, pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			case strings.Index("123456789", pos[0:1]) != -1:
				l = fmt.Sprintf("    %s"+barLine, pad(pos, 4))
				//l= fmt.Sprintf("    %4s+ |")
			case pos[0] == '.':
				l = fmt.Sprintf("     %s"+barLine, pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			case pos[0] == ';':
				l = fmt.Sprintf("     %s"+barLine, pad(pos, 3))
				//l = fmt.Sprintf("     %3s+ |")
			}

			for instrCol, instr := range p.score.Instruments {
				if ev, has := bar.events[barEventKey{instrCol, bar.positions[pi]}]; has && ev != nil {
					l += fmt.Sprintf(prefix+"%s"+barLine, instr.pad(ev.originalData))
				} else {
					l += fmt.Sprintf(prefix+"%s"+barLine, instr.pad(""))
				}
			}

			// fmt.Printf("bar: %v, pi: %v\n", bar.barNo, pi)

			if pi == 0 {
				for prt, br := range p.score.Parts {
					if br[0] == bar.originalBarNo {
						l += fmt.Sprintf(prefix+"%s", prt)
					}
				}
			}
			p.writeBodyLine(bf, l)
		}

	}

	lastSysLine := p.score.lastLine - p.score.lineWhereBodyStarts
	missingLines := lastSysLine - p.writerSysLine

	for i := 0; i < missingLines; i++ {
		if comment, has := p.score.BodyComments[p.writerSysLine+i]; has {
			bf.WriteString(comment + "\n")
		}
	}

}

func (p *Formatter) printHeader(bf *bytes.Buffer) {
	metaMap := map[string]string{}

	for k, v := range p.score.Meta {
		metaMap[pad("@"+k+":", 15)] = v
	}

	p.printSorted(bf, "%s %s\n", metaMap)

	bf.WriteString("\n\n")

	for _, incl := range p.score.HeaderIncludes {
		fmt.Fprintf(bf, "$include(%q)\n", incl)
	}

	p.printSorted(bf, "/%s/%s\n", p.score.Temperament)
	bf.WriteString("\n\n")

	pattDefMap := map[string]string{}

	for k, v := range p.score.PatternDefinitions {
		pattDefMap[pad(k+":", 15)] = v.Original
	}

	p.printSorted(bf, "%s %s\n", pattDefMap)

	bf.WriteString("\n\n")

	for _, comment := range p.score.HeaderComments {
		bf.WriteString(comment + "\n")
	}
}

func (p *Formatter) Format(bf *bytes.Buffer) {
	p.printHeader(bf)

	bf.WriteString("\n=\n")

	barLine := " |"
	if p.score.SmallColumns {
		barLine = "|"
	}

	p.printBody(bf, barLine)
}

func (p *Formatter) writeBodyLine(bf *bytes.Buffer, line string) {
	p._writeBodyLine(bf, line, true)
}

func (p *Formatter) _writeBodyLine(bf *bytes.Buffer, line string, lineBreak bool) {
	// we don't write comments for unrolled scores, since the lines won't match anyway
	if !p.score.isUnrolled {
		if comment, has := p.score.BodyComments[p.writerSysLine]; has {
			p.writerSysLine++
			bf.WriteString(comment + "\n")
			p._writeBodyLine(bf, line, lineBreak)
			return
		}
	}
	p.writerSysLine++
	if lineBreak {
		bf.WriteString(line + "\n")
		return
	}
	bf.WriteString(line)
}