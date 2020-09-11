package reference

import "testing"

func TestParse(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err bool
	}{
		{"'file=table.col[part]", "'file=table.col[part]", false},
		{"'file", "'file", false},
		{"'file=table", "'file=table", false},
		{"'file.table", "'file.table", false},
		{".table", ".table", false},
		{"=table", "=table", false},
		{"'file=table.col", "'file=table.col", false},
		{"'file.table.row", "'file.table.row", false},
		{"'file.table.row.col", "'file.table.row.col", false},
		{"=.table.col", "=.table.col", false},
		{"[part]", "[part]", false},

		{"'file=table.col[part", "", true},
		{"file", "file", true},
		{"'file'table.col[part]", "", true},
		{"=table'col[part]", "", true},
		{"=table.col['part]", "", true},
		{"=table=col[part]", "", true},
		{"=table.col[=part]", "", true},
		{".table=col[part]", "", true},
		{"]=table.col[part]", "", true},
		{"=table].col[part]", "", true},
		{"=table.col][part]", "", true},
		{"'file]=table.col[part]", "", true},
		{"'file.table'.col[part]", "", true},
		{".table'.col[part]", "", true},
		{".table].col[part]", "", true},
		{".table.col[[part]", "", true},
		{"=table.col[part]]", "", true},
	}

	for i, test := range tests {
		r, err := parse(test.in)

		if err != nil && !test.err {
			t.Errorf("[%v] parse(%#v) returns error: %#v", i, test.in, err.Error())
		}

		if err != nil {
			continue
		}

		if err == nil && test.err {
			t.Errorf("[%v] parse(%#v) should return error", i, test.in)
		}

		got, expected := r.String(), test.out

		if got != expected {
			t.Errorf("[%v] parse(%#v) = %#v // expected: %#v", i, test.in, got, expected)
		}

	}
}

func TestAll(t *testing.T) {
	var tests = []struct {
		ref      string
		ctx      string
		ctxTyp   Type
		err      bool
		typ      string
		valid    bool
		complete bool
		str      string
	}{
		// ok
		{"'file=table.col[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},
		{"'file=table[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},
		{"'file=table[part]", "'file=table.col1 col2 col3", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col1 col2 col3[part]"},
		{"'file[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},
		{"'file[part]", "'file=table.col1 col2 col3", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col1 col2 col3[part]"},
		{"=table.col[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},
		{"=.col[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},
		{"=.[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},
		{"[part]", "'file=table.col", ScoreColCtx, false, "ScoreColPart", true, true, "'file=table.col[part]"},

		{"'file=table.col2", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file=table.col2"},
		{"'file2=table", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file2=table.col"},
		{"'file=table2", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file=table2.col"},
		{"'file=table2", "'file=table.col1 col2 col3", ScoreColCtx, false, "ScoreCol", true, true, "'file=table2.col1 col2 col3"},
		{"'file2=", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file2=table.col"},
		{"'file2", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file2=table.col"},
		{"'file2", "'file=table.col1 col2 col3", ScoreColCtx, false, "ScoreCol", true, true, "'file2=table.col1 col2 col3"},
		{"=table.col2", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file=table.col2"},
		{"=.col2", "'file=table.col", ScoreColCtx, false, "ScoreCol", true, true, "'file=table.col2"},

		// errors
		{"'file=table.col", "'file=table.col", ScoreColCtx, true, "ScoreCol", false, false, ""},
		{"=table.col", "'file=table.col", ScoreColCtx, true, "ScoreCol", false, false, ""},
		{"=.col", "'file=table.col", ScoreColCtx, true, "ScoreCol", false, false, ""},

		// ok
		{"'file.table.row", "'file=table.col", ScoreColCtx, false, "ShortCutCell", true, true, "'file.table.row.col"},
		{".table.row", "'file=table.col", ScoreColCtx, false, "ShortCutCell", true, true, "'file.table.row.col"},

		// errors
		{".row", "'file=table.col", ScoreColCtx, true, "ShortCutCell", false, false, ""},

		// ok
		{"'file=table2", "'file=table", ScoreCtx, false, "Score", true, true, "'file=table2"},
		{"'file2", "'file=table", ScoreCtx, false, "Score", true, true, "'file2=table"},
		{"'file=table2[part]", "'file=table", ScoreCtx, false, "ScorePart", true, true, "'file=table2[part]"},
		{"'file2[part]", "'file=table", ScoreCtx, false, "ScorePart", true, true, "'file2=table[part]"},
		{"[part]", "'file=table", ScoreCtx, false, "ScorePart", true, true, "'file=table[part]"},
		{"=table2[part]", "'file=table", ScoreCtx, false, "ScorePart", true, true, "'file=table2[part]"},
		{"=table[part]", "'file=table", ScoreCtx, false, "ScorePart", true, true, "'file=table[part]"},

		// errors
		{"'file=table", "'file=table", ScoreCtx, true, "Score", false, false, ""},
		{"=table", "'file=table", ScoreCtx, true, "Score", false, false, ""},

		// ok
		{"'file2=table2", "'file", FileCtx, false, "Score", true, true, "'file2=table2"},
		{"'file2.table2", "'file", FileCtx, false, "ShortCut", true, true, "'file2.table2"},
		{"'file2", "'file", FileCtx, false, "File", true, true, "'file2"},

		// errors
		{"'file", "'file", FileCtx, true, "File", false, false, ""},
		{"'file=table", "'file", FileCtx, true, "Score", false, false, ""},
		{".table", "'file", FileCtx, true, "ShortCut", false, false, ""},

		// ok
		{"'file2.table.row.col", "'file.table", ShortCutCtx, false, "ShortCutCell", true, true, "'file2.table.row.col"},
		{"'file2.table.row", "'file.table", ShortCutCtx, false, "ShortCutCell", true, true, "'file2.table.row"},
		{"'file.table2.row", "'file.table", ShortCutCtx, false, "ShortCutCell", true, true, "'file.table2.row"},
		{"'file.table2.row.col", "'file.table", ShortCutCtx, false, "ShortCutCell", true, true, "'file.table2.row.col"},
		{".table2.row", "'file.table", ShortCutCtx, false, "ShortCutCell", true, true, "'file.table2.row"},
		{".table2.row.col", "'file.table", ShortCutCtx, false, "ShortCutCell", true, true, "'file.table2.row.col"},

		// errors
		{"'file2=table.col", "'file.table", ShortCutCtx, true, "", false, false, ""},
		{"=table.col", "'file.table", ShortCutCtx, true, "", false, false, ""},
		{".col", "'file.table", ShortCutCtx, true, "", false, false, ""},
	}

	for i, test := range tests {
		ctx, err2 := ParseContext(test.ctx)

		if err2 != nil {
			t.Errorf("[%v] error while parsing context %#v: %#v", i, test.ctx, err2.Error())
			continue
		}

		if ctx.Type.String() != test.ctxTyp.String() {
			t.Errorf("[%v] ParseContext(%#v).Type() = %#v // expected %#v", i, test.ctx, ctx.Type.String(), test.ctxTyp.String())
			continue
		}
		//		ctx.Type = test.ctxTyp

		r, err := Parse(test.ref, *ctx)

		if err != nil && !test.err {
			t.Errorf("[%v] Parse(%#v,%#v) returns error: %#v", i, test.ref, test.ctx, err.Error())
		}

		if err != nil {
			continue
		}

		if err == nil && test.err {
			t.Errorf("[%v] Parse(%#v,%#v) should return error", i, test.ref, test.ctx)
			continue
		}

		got, expected := r.Type.String(), test.typ

		if got != expected {
			t.Errorf("[%v] Parse(%#v,%#v) = %#v // expected: %#v", i, test.ref, test.ctx, got, expected)
		}

		if r.Valid() != test.valid {
			t.Errorf("[%v] Parse(%#v,%#v).Valid() = %v // expected: %v", i, test.ref, test.ctx, r.Valid(), test.valid)
		}

		if r.IsComplete() != test.complete {
			t.Errorf("[%v] Parse(%#v,%#v).IsComplete() = %v // expected: %v", i, test.ref, test.ctx, r.IsComplete(), test.complete)
		}

		if r.String() != test.str {
			t.Errorf("[%v] Parse(%#v,%#v).String() = %#v // expected: %#v", i, test.ref, test.ctx, r.String(), test.str)
		}
	}
}
