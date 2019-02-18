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
		instr.calcColWidth()
		//instrColWidths[i] = instr.colWidth
		if p.score.SmallColumns {
			l += fmt.Sprintf("%s|", instr.pad(instr.Name))
		} else {

			l += fmt.Sprintf(" %s |", instr.pad(instr.Name))
		}
	}

	p.writeBodyWithComment(bf, l)

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

	p.writeBodyWithComment(bf, l)

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

	p.writeBodyWithComment(bf, l)

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

	p.writeBodyWithComment(bf, l)

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

	p.writeBodyWithComment(bf, l)
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

// we don't write comments here, since the lines are changed anyway
func (p *Formatter) printUnrolled(bf *bytes.Buffer, barLine string) {
	var l string
	p.score.enroll()

	p.writeInstrumentLines(bf)

	//		fmt.Printf("len bars: %v\n", len(p.score.Bars))
	//		fmt.Printf("len events: %v\n", len(p.score.Instruments[0].events))

	for i, bar := range p.score.Bars {
		if bar.timeSigChange[0] != 0 {
			l = fmt.Sprintf("%v/%v", bar.timeSigChange[0], bar.timeSigChange[1])
		} else {
			l = ""
		}

		if bar.tempoChange > 0 {
			l += fmt.Sprintf("@%g", bar.tempoChange)
		}

		bf.WriteString(l + "\n")

		var instrPrinted = false

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

			for _, instr := range p.score.Instruments {
				instrPrinted = false
				be := instr.events[i]

				for _, iev := range be {
					if iev.DistanceToStartOfBarIn32th == bar.positions[pi] {
						if p.score.SmallColumns {
							l += fmt.Sprintf("%s"+barLine, instr.pad(iev.originalData))
						} else {
							l += fmt.Sprintf(" %s"+barLine, instr.pad(iev.originalData))
						}
						instrPrinted = true
						break
					}

					if iev.DistanceToStartOfBarIn32th > bar.positions[pi] {
						if p.score.SmallColumns {
							l += fmt.Sprintf("%s"+barLine, instr.pad(""))
						} else {
							l += fmt.Sprintf(" %s"+barLine, instr.pad(""))
						}
						instrPrinted = true
						break
					}
				}

				if !instrPrinted {
					if p.score.SmallColumns {
						l += fmt.Sprintf("%s"+barLine, instr.pad(""))
					} else {
						l += fmt.Sprintf(" %s"+barLine, instr.pad(""))
					}
				}
			}

			if pi == 0 {
				for prt, br := range p.score.Parts {
					if br[0] == bar.originalBarNo {
						l += fmt.Sprintf(" %s", prt)
					}
				}
			}

			bf.WriteString(l + "\n")

		}

	}
}

func (p *Formatter) printNormalBody(bf *bytes.Buffer, barLine string) {
	var l string

	p.writeInstrumentLines(bf)

	//		fmt.Printf("len bars: %v\n", len(p.score.Bars))

	for i, bar := range p.score.Bars {
		if bar.jumpTo != "" {
			l = fmt.Sprintf("[%s]", bar.jumpTo)
			p.writeBodyWithComment(bf, l)
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

		p.writeBodyWithComment(bf, l)
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

			for _, instr := range p.score.Instruments {
				be := instr.events[i]

				l += fmt.Sprintf(" %s"+barLine, instr.pad(be[pi].originalData))

			}

			if pi == 0 {
				for prt, br := range p.score.Parts {
					if br[0] == i {
						if p.score.SmallColumns {
							l += fmt.Sprintf("%s", prt)

						} else {
							l += fmt.Sprintf(" %s", prt)
						}
					}
				}
			}

			/*
				if pi == len(bar.originalPositions)-1 {
					for br, jmp := range p.score.Jumps {
						if br == i {
							l += fmt.Sprintf(" %s", jmp)
						}
					}
				}
			*/
			p.writeBodyWithComment(bf, l)
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

	if p.score.isUnrolled {
		p.printUnrolled(bf, barLine)
		return
	}

	p.printNormalBody(bf, barLine)
}

func (p *Formatter) writeBodyWithComment(bf *bytes.Buffer, line string) {
	if comment, has := p.score.BodyComments[p.writerSysLine]; has {
		p.writerSysLine++
		bf.WriteString(comment + "\n")
		p.writeBodyWithComment(bf, line)
		return
	}
	p.writerSysLine++
	bf.WriteString(line + "\n")
}
