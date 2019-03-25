package muskel_test

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/formatter"
	"gitlab.com/gomidi/muskel/unroller"
)

func TestFormatterJumps(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`

=
         | <Vocals> |
    1    | e      | A
[A]
    2    | f      |
`,
			`

=
         | <Vocals> |
    1    | e        | A
[A]
    2    | f        |
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      |  A
*2
    2    | f      |
`,
			`

=
         | <Vocals> |
    1    | e        | A
*2
    2    | f        |
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      | A
[A]
[A]
    2    | f      |
`,
			`

=
         | <Vocals> |
    1    | e        | A
[A]
[A]
    2    | f        |
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      | A
[A]
[A]
3/4
    2    | f      |
`,
			`

=
         | <Vocals> |
    1    | e        | A
[A]
[A]
3/4
    2    | f        |
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      | A
[A]
[A]
3/4@120
    2    | f      |
`,
			`

=
         | <Vocals> |
    1    | e        | A
[A]
[A]
3/4@120
    2    | f        |
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      | 
3/4@120~
    2    | f      |
@100
`,
			`

=
         | <Vocals> |
    1    | e        |
3/4@120~
    2    | f        |
@100
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      | 
@120~~\major^d
    2    | f      |
@100
`,
			`

=
         | <Vocals> |
    1    | e        |
@120~~\major^d
    2    | f        |
@100
`,
		},
		{
			`

=
         | <Vocals> |
    1    | e      | A
[A]
[A]
@120
    2    | f      |
`,
			`

=
         | <Vocals> |
    1    | e        | A
[A]
[A]
@120
    2    | f        |
`,
		},
	}

	for i, test := range tests {
		/*
			if i != 1 {
				continue
			}
		*/
		sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "formatter   jumps")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder

		fm := formatter.New(sc)
		fm.HideInstrumentProperties = true
		fm.WriteTo(&bf)

		result := strings.TrimSpace(bf.String())
		expected := strings.TrimSpace(test.expected)

		//		fmt.Println(result)

		if result != expected {
			t.Errorf("[%v] score\n%s\n\nformatted gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, result, expected, result, expected)
			//			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}

func TestFormatter(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
@prop: this is a: value
aa:             1f" 2&g

=
         | <Vocals> |
Ch       | 1      |
Bank     |          |
Prog     |          |
Vol      |          |
    1    | e'     |
    2    | aa     |

    1    |./.     |

    1    | %      |`,
			`
@prop:          this is a: value




aa:             1f" 2&g



=
        |<Vocals>|
File    |        |
Ch      |1       |
Bank    |        |
Prog    |        |
Vol     |        |
PbRange |        |
Trans   |        |
Delay   |        |
    1   |e'      |
    2   |f"      |
    3&  |g       |

    1   |e'      |
    2   |f"      |
    3&  |g       |

    1   |g       |`,
		},
		{
			`
aa:             1f" 2&g

=
         | <Vocals> |
Ch       | 1        |
Bank     |          |
Prog     |          |
Vol      |          |
    1    | e'     |
    2    | aa     |
*3
    1    |f"     |`,
			`
aa:             1f" 2&g



=
        |<Vocals>|
File    |        |
Ch      |1       |
Bank    |        |
Prog    |        |
Vol     |        |
PbRange |        |
Trans   |        |
Delay   |        |
    1   |e'      |
    2   |f"      |
    3&  |g       |
*3
    1   |f"      |`,
		},
		{
			`
@author: Heino
@title: hoch auf dem gelben wagen

/a"/440Hz

aa: 1a' 2&#1 5&c"
bb: 1aa(#1) 5aa(d)
=
  | piano | vox |
1 | aa(g) | bb(f)   |

1 | ./.   |    |
`, `
@author:        Heino
@title:         hoch auf dem gelben wagen


/a"/440Hz


aa:             1a' 2&#1 5&c"
bb:             1aa(#1) 5aa(d)



=
        |<piano>|<vox>|
File    |       |     |
Ch      |       |     |
Bank    |       |     |
Prog    |       |     |
Vol     |       |     |
PbRange |       |     |
Trans   |       |     |
Delay   |       |     |
    1   |a'     |a'   |
    2&  |g      |f    |

    1   |a'     |a'   |
    2&  |g      |d    |
			`},
	}

	for i, test := range tests {
		/*
			if i != 0 {
				continue
			}
		*/
		sc, err := muskel.Parse(strings.NewReader(strings.TrimSpace(test.input)), "formatter")

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}
		sc.SmallColumns = true

		unr, err := unroller.Unroll(sc)

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf strings.Builder

		err = muskel.WriteFormattedTo(unr, &bf)

		if err != nil {
			t.Errorf("[%v] could not format unrolled score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		result := strings.TrimSpace(bf.String())
		expected := strings.TrimSpace(test.expected)

		//		fmt.Println(result)

		if result != expected {
			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%s\n\nbut this was expected:\n%s\n%q\nvs\n%q\n", i, test.input, result, expected, result, expected)
			//			t.Errorf("[%v] score\n%s\n\nunrolled gives \n%q\n\nbut this was expected:\n%q\n", i, test.input, got, wanted)
		}
	}
}
