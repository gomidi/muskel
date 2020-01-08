package table

type Timbres struct {
	*Table
}

func NewTimbres(lineNo int, sc Score) *Timbres {
	t := &Timbres{
		Table: NewTable("TIMBRE", lineNo, sc),
	}
	//sc.timbres = t
	return t
}
