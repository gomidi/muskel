package muskel

import (
	"bytes"
	"fmt"
	"strings"
)

type Formatter struct {
	score         *Score
	writerSysLine int
}

func NewFormatter(s *Score) *Formatter {
	return &Formatter{score: s}
}

func (p *Formatter) Format(bf *bytes.Buffer) {
	//bf.WriteString("\n// info\n")

	//fmt.Fprintf(bf, "%s %s\n", pad("version:", 15), time.Now().Format(time.RFC822))

	for k, v := range p.score.Meta {
		fmt.Fprintf(bf, "%s %s\n", pad(k+":", 15), v)
		//fmt.Fprintf(bf, "%15s+ %s\n", k+":", v)
	}

	//bf.WriteString("\n// temperament\n")

	bf.WriteString("\n\n")

	for k, v := range p.score.Temperament {
		fmt.Fprintf(bf, "/%s/%s\n", k, v)
	}

	bf.WriteString("\n\n")

	//bf.WriteString("\n// pattern definitions\n")

	for k, v := range p.score.PatternDefinitions {
		//fmt.Fprintf(bf, "%s %s\n", pad("$"+k+":", 15), v)
		fmt.Fprintf(bf, "%s %s\n", pad("$"+k+":", 15), v.Original)
		//fmt.Fprintf(bf, "%15s+ %s\n", "$"+k+":", v.Original)
	}

	bf.WriteString("\n\n")

	for _, comment := range p.score.HeaderComments {
		bf.WriteString(comment + "\n")
	}

	//fmt.Fprintf(bf, "\n\nBPM: %0.2f\n", p.tempoBPM)

	bf.WriteString("\n=\n")

	// 9 whitespace to first pipe
	l := "         |"

	for _, instr := range p.score.Instruments {
		instr.calcColWidth()
		//instrColWidths[i] = instr.colWidth
		l += fmt.Sprintf(" %s |", instr.pad(instr.Name))
	}

	p.writeSystemLine(bf, l)

	l = "Ch       |"

	for _, instr := range p.score.Instruments {
		if instr.MIDIChannel == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIChannel+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Bank     |"

	for _, instr := range p.score.Instruments {
		if instr.MIDIBank == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIBank+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Prog     |"

	for _, instr := range p.score.Instruments {
		if instr.MIDIProgram == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIProgram+1)))
		}
	}

	p.writeSystemLine(bf, l)

	l = "Vol      |"

	for _, instr := range p.score.Instruments {
		if instr.MIDIVolume == -1 {
			l += fmt.Sprintf(" %v |", instr.pad("-"))
		} else {
			l += fmt.Sprintf(" %v |", instr.pad(fmt.Sprintf("%v", instr.MIDIVolume+1)))
		}
	}

	p.writeSystemLine(bf, l)

	if p.score.isUnrolled {

		// enroll the instrument events again
		for _, instr := range p.score.Instruments {
			instr.events = []BarEvents{}
			var be BarEvents
			var lastBarNo int

			for _, ev := range instr.unrolled {
				p.score.Bars[ev.BarNo].ensurePositionExist(ev.DistanceToStartOfBarIn32th)

				if ev.BarNo > lastBarNo {
					instr.events = append(instr.events, be)
					be = BarEvents{}
					lastBarNo = ev.BarNo
				}

				be = append(be, ev)
			}

			instr.events = append(instr.events, be)

			instr.calcColWidth()
		}

		for i, bar := range p.score.Bars {
			if bar.timeSigChange[0] != 0 {
				l = fmt.Sprintf("%v/%v", bar.timeSigChange[0], bar.timeSigChange[1])
			} else {
				l = ""
			}

			if bar.tempoChange > 0 {
				l += fmt.Sprintf("@%g", bar.tempoChange)
			}

			p.writeSystemLine(bf, l)

			for pi, pos := range bar.originalPositions {
				switch {
				case pos[0] == '&':
					l = fmt.Sprintf("     %s |", pad(pos, 3))
					//l = fmt.Sprintf("     %3s+ |")
				case strings.Index("123456789", pos[0:1]) != -1:
					l = fmt.Sprintf("    %s |", pad(pos, 4))
					//l= fmt.Sprintf("    %4s+ |")
				case pos[0] == '.':
					l = fmt.Sprintf("     %s |", pad(pos, 3))
					//l = fmt.Sprintf("     %3s+ |")
				case pos[0] == ';':
					l = fmt.Sprintf("     %s |", pad(pos, 3))
					//l = fmt.Sprintf("     %3s+ |")
				}

				for _, instr := range p.score.Instruments {
					be := instr.events[i]

					for _, iev := range be {
						if iev.DistanceToStartOfBarIn32th == bar.positions[pi] {
							l += fmt.Sprintf(" %s |", instr.pad(be[pi].originalData))
							break
						}

						if iev.DistanceToStartOfBarIn32th > bar.positions[pi] {
							l += fmt.Sprintf(" %s |", instr.pad(""))
							break
						}
					}
				}

				/*
					if pi == 0 {
						for prt, br := range p.score.Parts {
							if br == i {
								l += fmt.Sprintf(" %s", prt)
							}
						}
					}
				*/
				/*
					if pi == len(bar.originalPositions)-1 {
						for br, jmp := range p.score.Jumps {
							if br == i {
								l += fmt.Sprintf(" %s", jmp)
							}
						}
					}
				*/
				p.writeSystemLine(bf, l)
			}

		}
	} else {

		for i, bar := range p.score.Bars {
			if bar.timeSigChange[0] != 0 {
				l = fmt.Sprintf("%v/%v", bar.timeSigChange[0], bar.timeSigChange[1])
			} else {
				l = ""
			}

			if bar.tempoChange > 0 {
				l += fmt.Sprintf("@%g", bar.tempoChange)
			}

			p.writeSystemLine(bf, l)

			for pi, pos := range bar.originalPositions {
				switch {
				case pos[0] == '&':
					l = fmt.Sprintf("     %s |", pad(pos, 3))
					//l = fmt.Sprintf("     %3s+ |")
				case strings.Index("123456789", pos[0:1]) != -1:
					l = fmt.Sprintf("    %s |", pad(pos, 4))
					//l= fmt.Sprintf("    %4s+ |")
				case pos[0] == '.':
					l = fmt.Sprintf("     %s |", pad(pos, 3))
					//l = fmt.Sprintf("     %3s+ |")
				case pos[0] == ';':
					l = fmt.Sprintf("     %s |", pad(pos, 3))
					//l = fmt.Sprintf("     %3s+ |")
				}

				for _, instr := range p.score.Instruments {
					be := instr.events[i]

					l += fmt.Sprintf(" %s |", instr.pad(be[pi].originalData))

				}

				if pi == 0 {
					for prt, br := range p.score.Parts {
						if br == i {
							l += fmt.Sprintf(" %s", prt)
						}
					}
				}

				if pi == len(bar.originalPositions)-1 {
					for br, jmp := range p.score.Jumps {
						if br == i {
							l += fmt.Sprintf(" %s", jmp)
						}
					}
				}

				p.writeSystemLine(bf, l)
			}

		}
	}
	//bf.WriteString("// END\n")
	bf.WriteString("\n")
}

func (p *Formatter) writeSystemLine(bf *bytes.Buffer, line string) {
	if comment, has := p.score.Comments[p.writerSysLine]; has {
		p.writerSysLine++
		bf.WriteString(comment + "\n")
		p.writeSystemLine(bf, line)
		return
	}
	p.writerSysLine++
	bf.WriteString(line + "\n")
}
