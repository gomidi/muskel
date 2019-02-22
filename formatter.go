package muskel

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type Formatter struct {
	score         *Score
	writerSysLine int
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
		//		f.score.Bars = f.score.un
	}

	for _, b := range f.score.Bars {
		b.events = map[barEventKey]*Event{}

		if f.score.isUnrolled {
			for i, instr := range f.score.Instruments {
				for _, ev := range instr.unrolled {
					if ev.BarNo == b.barNo {
						b.addInstrEvent(ev.DistanceToStartOfBarIn32th, i, ev)
					}
				}
			}
		} else {
			for i, instr := range f.score.Instruments {
				for j, bev := range instr.events {
					if j == b.barNo {
						for bidx, ev := range bev {
							//b.addInstrEvent(position, i, ev)
							b.addInstrEvent(b.positions[bidx], i, ev)
							//							b.events[barEventKey{instrCol: i, position: b.positions[bidx]}] = ev
						}
					}
				}
			}
		}

		/*
			for i, pos := range b.positions {
				if b.originalPositions[i] != pos32thToString(pos) {
					panic(fmt.Sprintf("strange things happening in bar %v at pos %v: %v 32th distance is not %q, but %q", b.barNo, i, pos, b.originalPositions[i], pos32thToString(pos)))
				}
				b.originalPositions[i] = pos32thToString(pos)
			}
		*/
	}
}

func (p *Formatter) printBody(bf *bytes.Buffer, barLine string) {
	var l string

	p.writeInstrumentLines(bf)
	p.prepareBars()

	var prefix string

	if !p.score.SmallColumns {
		prefix = " "
	}

	for _, bar := range p.score.Bars {

		if bar.jumpTo != "" {
			l = fmt.Sprintf("[%s]", bar.jumpTo)
			p.writeBodyLine(bf, l)
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

		p.writeBodyLine(bf, l)
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
		metaMap[pad(k+":", 15)] = v
	}

	p.printSorted(bf, "%s %s\n", metaMap)

	bf.WriteString("\n\n")

	p.printSorted(bf, "/%s/%s\n", p.score.Temperament)
	bf.WriteString("\n\n")

	pattDefMap := map[string]string{}

	for k, v := range p.score.PatternDefinitions {
		pattDefMap[pad("$"+k+":", 15)] = v.Original
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
	// we don't write comments for unrolled scores, since the lines won't match anyway
	if !p.score.isUnrolled {
		if comment, has := p.score.BodyComments[p.writerSysLine]; has {
			p.writerSysLine++
			bf.WriteString(comment + "\n")
			p.writeBodyLine(bf, line)
			return
		}
	}
	p.writerSysLine++
	bf.WriteString(line + "\n")
}
