package muskel

type barEventKey struct {
	instrCol int
	position uint
}

type Bar struct {
	tempoChange       float64
	timeSigChange     [2]uint8
	positions         []uint   // each position is number of 32th from start
	originalPositions []string // original positioning strings
	lineNo            int      // line numbers are counted from the start of the system
	barNo             int
	originalBarNo     int
	// real time sig (inherited or changed)
	timeSig       [2]uint8
	jumpTo        string
	include       string
	includedScore *Score // must be an unrolled score
	events        map[barEventKey]*Event
}

func NewBar() *Bar {
	return &Bar{
		events: map[barEventKey]*Event{},
	}
}

func (b *Bar) addInstrEvent(position uint, instrCol int, ev *Event) {
	b.ensurePositionExist(position)
	b.events[barEventKey{instrCol: instrCol, position: position}] = ev
}

func (b *Bar) ensurePositionExist(pos uint) {

	for i, p := range b.positions {
		if p == pos {
			return
		}

		if p > pos {
			np := append([]uint{}, b.positions[:i]...)
			np = append(np, pos)
			np = append(np, b.positions[i:]...)
			b.positions = np

			op := append([]string{}, b.originalPositions[:i]...)
			op = append(op, pos32thToString(pos))
			op = append(op, b.originalPositions[i:]...)
			b.originalPositions = op
			return
		}
	}

	b.positions = append(b.positions, pos)
	b.originalPositions = append(b.originalPositions, pos32thToString(pos))
	return
}

func (b *Bar) length32th() int {
	return length32ths(b.timeSig[0], b.timeSig[1])
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
		tempoChange:       b.tempoChange,
		timeSigChange:     b.timeSigChange,
		positions:         b.positions,
		originalPositions: b.originalPositions,
		lineNo:            b.lineNo,
		barNo:             b.barNo,
		originalBarNo:     b.originalBarNo,
		timeSig:           b.timeSig,
		jumpTo:            b.jumpTo,
		include:           b.include,
		events:            map[barEventKey]*Event{},
		includedScore:     b.includedScore,
	}
}

type Event struct {
	Item                       interface{}
	originalData               string
	BarNo                      int
	OriginalBarNo              int
	DistanceToStartOfBarIn32th uint
	lineNo                     int // line numbers are counted from the start of the system
}

func (e *Event) Dup() *Event {
	return &Event{
		Item:                       e.Item,
		originalData:               e.originalData,
		BarNo:                      e.BarNo,
		OriginalBarNo:              e.OriginalBarNo,
		DistanceToStartOfBarIn32th: e.DistanceToStartOfBarIn32th,
		lineNo:                     e.lineNo,
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

func (b BarEvents) isEmpty() bool {
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
		_, is := ev.Item.(hold)
		if is {
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
		case RepeatLastBar:
			return 1, false
		case RepeatLastBarUntilChange:
			return 1, true
		case RepeatLastNBarsUntilChange:
			return int(v), true
		}
	}

	return
}
