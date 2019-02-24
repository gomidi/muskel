package muskel

import (
	"strings"
	"testing"
)

func TestParserBarLength(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			`
=
   | Voc | Drums | Piano |
Ch | -   | -     | -     |
1  | a"  | d"   | b"    |
2& | c   |     | d     |

1  |     |     |       |
`,
			2,
		},
		{
			`
=
   | Voc | Drums | Piano |
Ch | -   | -     | -     |
4/4
1  | a"  | d    | b"    |
2& | c   |      | d     |

1  | d   |      |       |
`,
			2,
		},
		{
			`
=
   | Voc | Drums | Piano |
Ch | -   | -     | -     |
4/4@140
1  | a"  | d    | b"    |
2& | c   |      | d     |

1  | d   |      |       |
`,
			2,
		},
		{
			`
=
   | Voc | Drums | Piano |
Ch | -   | -     | -     |
@100
1  | a"  | d    | b"    |
2& | c   |      | d     |

1  | d   |      |       |
`,
			2,
		},
	}

	for i, test := range tests {

		if i != 1 {
			continue
		}

		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)), "include-main")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		unr, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := len(unr.Bars)

		got := result
		wanted := test.expected

		if got != wanted {
			t.Errorf("[%v] score\n%s\n\nunrolled gives %v bars // expected: %v", i, test.input, got, wanted)
		}
	}
}
