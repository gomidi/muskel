package muskel

type Bar struct {
	tempoChange       float64
	timeSigChange     [2]uint8
	positions         []uint8  // each position is number of 32th from start
	originalPositions []string // original positioning strings
}

type Event struct {
	Item         interface{}
	originalData string
}

type BarEvents []*Event

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
