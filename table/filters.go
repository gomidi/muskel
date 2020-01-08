package table

type Filters struct {
	*Table
}

func NewFilters(lineNo int, sc Score) *Filters {
	t := &Filters{
		Table: NewTable("FILTER", lineNo, sc),
	}
	//sc.filters = t
	return t
}
