package muskel_test

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/formatter"
	"gitlab.com/gomidi/muskel/unroller"
)

func TestRandom2(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected [2]string
	}{
		{
			`
=
  | <piano> |
1 | ?(a,b) | 

`, [2]string{`
         | <piano> |
    1    | a       |
`,
				`
         | <piano> |
    1    | b       |
`,
			},
		},
		{
			`
aa: 1c
bb: 2d
=
  | <piano> |
1 | ?(aa,bb) | 

`, [2]string{`
         | <piano> |
    1    | c       |
`,
				`
         | <piano> |
    1    |         |
    2    | d       |
`,
			},
		},
		{
			`
=
  | <piano> |
1 | ?50%a | 

`, [2]string{`
         | <piano> |
    1    | a       |
`,
				`
         | <piano> |
    1    |         |
`,
			},
		},
	}

	for i, test := range tests {

		//		if i != 0 {
		//			continue
		//		}

		sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "random2")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}
		//		sc.SmallColumns = true

		unr, err := unroller.Unroll(sc)

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder
		fm := formatter.New(unr)
		fm.HideTrackProperties = true
		fm.HideHeader = true
		fm.WriteTo(&bf)

		result := strings.TrimSpace(bf.String())

		expected := test.expected

		expected[0] = strings.TrimSpace(test.expected[0])
		expected[1] = strings.TrimSpace(test.expected[1])

		if result != expected[0] && result != expected[1] {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut one of these were expected:\n%s\n%s\n", i, test.input, result, expected[0], expected[1])
			t.Errorf("[%v] score\n%q\n\nunrolled gives \n%q\n\nbut one of these were expected:\n%q\n%q\n", i, test.input, result, expected[0], expected[1])
		}
	}
}

func TestRandom1(t *testing.T) {
	//t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
=
  | <piano> | <vox> |
1 | ?100%a | ?0%c   |

1 | ./.   |    |
`, `
=
         | <piano> | <vox> |
    1    | a       |       |

    1    | a       |       |
`},
		{
			`
=
  | <piano> |
1 | ?(b) | 

`, `
=
         | <piano> |
    1    | b       |
`},
	}

	for i, test := range tests {

		//		if i != 0 {
		//			continue
		//		}

		sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "random1")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}
		//		sc.SmallColumns = true

		unr, err := unroller.Unroll(sc)

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder
		fm := formatter.New(unr)
		fm.HideTrackProperties = true
		fm.WriteTo(&bf)

		result := strings.TrimSpace(bf.String())
		expected := strings.TrimSpace(test.expected)

		if result != expected {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, result, expected, result, expected)
		}
	}
}
