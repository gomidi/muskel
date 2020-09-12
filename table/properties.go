package table

type Properties struct {
	*Table
}

func NewProperties(lineNo int, sc Score) *Properties {
	t := &Properties{
		Table: NewTable("PROPERTY", lineNo, sc),
	}
	return t
}

func (t *Properties) Finish() (err error) {
	for _, line := range t.Table.Data {
		t.Table.Score.AddProperty(line[0], line[1])
	}

	t.cols = []string{""}
	return nil
}
