package muskel

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatter(t *testing.T) {
	//	t.Skip()
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
$a:             1f" 2&g

=
         | Vocals |
Ch       | 1      |
Bank     | -      |
Prog     | -      |
Vol      | -      |

    1    | e'     |
    2    | $a     |

    1    |./.     |

    1    | %      |`,
			`
$a:             1f" 2&g



=
        |Vocals|
Ch      |1     |
Bank    |-     |
Prog    |-     |
Vol     |-     |

    1   |e'    |
    2   |f"    |
    3&  |g     |

    1   |e'    |
    2   |f"    |
    3&  |g     |

    1   |g     |`,
		},
		{
			`
author: Heino
title: hoch auf dem gelben wagen

/a"/440Hz

$a: 1a' 2&#1 5&c"
$b: 1$a(#1) 5$a(d)
=
  | piano | vox |

1 | $a(g) | $b(f)   |

1 | ./.   |    |
`, `
author:         Heino
title:          hoch auf dem gelben wagen


/a"/440Hz


$a:             1a' 2&#1 5&c"
$b:             1$a(#1) 5$a(d)



=
        |piano|vox|
Ch      |-    |-  |
Bank    |-    |-  |
Prog    |-    |-  |
Vol     |-    |-  |

    1   |a'   |a' |
    2&  |g    |f  |

    1   |a'   |a' |
    2&  |g    |d  |
			`},
	}

	for i, test := range tests {

		//		if i != 0 {
		//			continue
		//		}

		sc, err := Parse(strings.NewReader(strings.TrimSpace(test.input)))

		if err != nil {
			t.Errorf("[%v] could not parse score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}
		sc.SmallColumns = true

		unr, err := sc.Unroll()

		if err != nil {
			t.Errorf("[%v] could not unroll score: %s\n%s\n", i, err.Error(), test.input)
			continue
		}

		var bf bytes.Buffer

		err = unr.WriteTo(&bf)

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
