package score

import "gitlab.com/gomidi/muskel/items"

type BarEventKey struct {
	InstrCol int
	Position uint
}

type Bar struct {
	TempoChange       float64
	TimeSigChange     [2]uint8
	Positions         []uint   // each position is number of 32th from start
	OriginalPositions []string // original positioning strings
	LineNo            int      // line numbers are counted from the start of the system
	BarNo             int
	OriginalBarNo     int
	// real time sig (inherited or changed)
	TimeSig       [2]uint8
	JumpTo        string
	Include       string
	IncludedScore *Score // must be an unrolled score
	Events        map[BarEventKey]*Event
	Scale         *items.Scale
	IsEmpty       bool
	Tilde         string
}

func NewBar() *Bar {
	return &Bar{
		Events: map[BarEventKey]*Event{},
	}
}

func (b *Bar) AddInstrEvent(position uint, instrCol int, ev *Event) {
	b.ensurePositionExist(position)
	b.Events[BarEventKey{InstrCol: instrCol, Position: position}] = ev
}

func (b *Bar) ensurePositionExist(pos uint) {

	for i, p := range b.Positions {
		if p == pos {
			return
		}

		if p > pos {
			np := append([]uint{}, b.Positions[:i]...)
			np = append(np, pos)
			np = append(np, b.Positions[i:]...)
			b.Positions = np

			op := append([]string{}, b.OriginalPositions[:i]...)
			op = append(op, items.Pos32thToString(pos))
			op = append(op, b.OriginalPositions[i:]...)
			b.OriginalPositions = op
			return
		}
	}

	b.Positions = append(b.Positions, pos)
	b.OriginalPositions = append(b.OriginalPositions, items.Pos32thToString(pos))
	return
}

func (b *Bar) length32th() int {
	return items.Length32ths(b.TimeSig[0], b.TimeSig[1])
}

func (b *Bar) Dup() (nuB *Bar) {
	/*
		tempoChange       float64
		timeSigChange     [2]uint8
		positions         []uint   // each position is number of 32th from start
		originalPositions []string // original positioning strings
		lineNo            int      // line numbers are counted from the start of the system
		barNo             int
		originalBarNo     int
		// real time sig (inherited or changed)
		timeSig [2]uint8
		jumpTo  string
		events  map[barEventKey]*Event
	*/
	return &Bar{
		TempoChange:       b.TempoChange,
		TimeSigChange:     b.TimeSigChange,
		Positions:         b.Positions,
		OriginalPositions: b.OriginalPositions,
		LineNo:            b.LineNo,
		BarNo:             b.BarNo,
		OriginalBarNo:     b.OriginalBarNo,
		TimeSig:           b.TimeSig,
		JumpTo:            b.JumpTo,
		Include:           b.Include,
		Events:            map[BarEventKey]*Event{},
		IncludedScore:     b.IncludedScore,
		Scale:             b.Scale,
		IsEmpty:           b.IsEmpty,
		Tilde:             b.Tilde,
	}
}

type Event struct {
	Item                       interface{}
	OriginalData               string
	BarNo                      int
	OriginalBarNo              int
	DistanceToStartOfBarIn32th uint
	LineNo                     int // line numbers are counted from the start of the system
}

func (e *Event) Dup() *Event {
	return &Event{
		Item:                       e.Item,
		OriginalData:               e.OriginalData,
		BarNo:                      e.BarNo,
		OriginalBarNo:              e.OriginalBarNo,
		DistanceToStartOfBarIn32th: e.DistanceToStartOfBarIn32th,
		LineNo:                     e.LineNo,
	}
}

type BarEvents []*Event

func (b BarEvents) Dup() BarEvents {
	var nu BarEvents
	for _, be := range b {
		nu = append(nu, be.Dup())
	}
	return nu
}

func (b BarEvents) IsEmpty() bool {
	if len(b) == 0 {
		return true
	}

	for _, e := range b {
		if e.Item != nil {
			return false
		}
	}

	return true
}

func (be BarEvents) isHolding() bool {
	if len(be) == 0 {
		return false
	}

	for _, ev := range be {
		// _, is := ev.Item.(hold)
		if ev.Item == items.Hold {
			return true
		}
	}
	return false
}

func (be BarEvents) RepeatingBars() (num int, untilNext bool) {
	if len(be) == 0 {
		return
	}

	for _, ev := range be {
		switch v := ev.Item.(type) {
		case items.RepeatLastBar:
			return 1, false
		case items.RepeatLastBarUntilChange:
			return 1, true
		case items.RepeatLastNBarsUntilChange:
			return int(v), true
		}
	}

	return
}
