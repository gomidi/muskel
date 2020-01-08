package table

type Tunings struct {
	*Table
}

func NewTunings(lineNo int, sc Score) *Tunings {
	t := &Tunings{
		Table: NewTable("TUNING", lineNo, sc),
	}
	//	sc.tunings = t
	return t
}
