package score

type Instrument struct {
	Name           string
	Delay          [2]int // fraction of 4/4
	MIDIChannel    int8
	MIDIProgram    int8
	MIDIVolume     int8
	MIDIBank       int8
	MIDITranspose  int8 // only transposes Notes not MIDINotes
	PitchbendRange uint8
	Events         []BarEvents // in the order of bars
	colWidth       int
	Unrolled       []*Event
	FileGroup      string
}

func (i *Instrument) Dup() *Instrument {
	return &Instrument{
		Name:           i.Name,
		Delay:          i.Delay,
		MIDIChannel:    i.MIDIChannel,
		MIDIProgram:    i.MIDIProgram,
		MIDIVolume:     i.MIDIVolume,
		MIDIBank:       i.MIDIBank,
		PitchbendRange: i.PitchbendRange,
		MIDITranspose:  i.MIDITranspose,
		colWidth:       i.colWidth, // ? set to 0??
		Events:         i.Events,
		FileGroup:      i.FileGroup,
	}
}

func (i *Instrument) Pad(s string) string {
	return Pad(s, i.colWidth)
}

func (i *Instrument) CalcColWidth(unrolled bool) {
	i.colWidth = len(i.Name) + 2

	if unrolled {
		for _, ev := range i.Unrolled {
			l := len(ev.OriginalData)
			if l > i.colWidth {
				i.colWidth = l
			}
		}
	} else {
		for _, be := range i.Events {
			for _, ev := range be {
				l := len(ev.OriginalData)
				if l > i.colWidth {
					i.colWidth = l
				}
			}
		}
	}

	if i.colWidth < 2 {
		i.colWidth = 2
	}
}
