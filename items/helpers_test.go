package items

import (
	"reflect"
	"testing"
)

func TestSplitParams(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a,{b,c}", []string{"a", "{b,c}"}},
		{"a,{b,c},e", []string{"a", "{b,c}", "e"}},
		{"a,(b c),e", []string{"a", "(b c)", "e"}},
		{"a,(b c)...,e", []string{"a", "(b c)...", "e"}},
		{"a,{b,{g,b}},e", []string{"a", "{b,{g,b}}", "e"}},
	}

	for i, test := range tests {

		result := splitParams(test.input)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%v] splitParams(%v) = %#v // expected: %#v", i, test.input, result, test.expected)
		}
	}
}

func TestEuclideanRhythm1(t *testing.T) {
	tests := []struct {
		start32th uint
		n         string
		m         string
		dur       string
		expected  [][2]string
	}{
		{0, "3", "8", "&",
			[][2]string{
				{"#", ""},
				{"1", "#1"},
				{"1&", "#2"},
				{"2", "#2"},
				{"2&", "#1"},
				{"3", "#2"},
				{"3&", "#2"},
				{"4", "#1"},
				{"4&", "#2"},
			},
		},
		{0, "3", "8", "1",
			[][2]string{
				{"#", ""},
				{"1", "#1"},
				{"2", "#2"},
				{"3", "#2"},
				{"4", "#1"},
				{"#", ""},
				{"1", "#2"},
				{"2", "#2"},
				{"3", "#1"},
				{"4", "#2"},
			},
		},
		{0, "3", "8", ",",
			[][2]string{
				{"#", ""},
				{"1", "#1"},
				{"1,", "#2"},
				{"1&", "#2"},
				{"1&,", "#1"},
				{"2", "#2"},
				{"2,", "#2"},
				{"2&", "#1"},
				{"2&,", "#2"},
			},
		},
		{8, "3", "8", "&",
			[][2]string{
				{"#", ""},
				{"2", "#1"},
				{"2&", "#2"},
				{"3", "#2"},
				{"3&", "#1"},
				{"4", "#2"},
				{"4&", "#2"},
				{"#", ""},
				{"1", "#1"},
				{"1&", "#2"},
			},
		},
		{4, "5", "8", "&",
			[][2]string{
				{"#", ""},
				{"1&", "#1"},
				{"2", "#2"},
				{"2&", "#1"},
				{"3", "#1"},
				{"3&", "#2"},
				{"4", "#1"},
				{"4&", "#1"},
				{"#", ""},
				{"1", "#2"},
			},
		},
	}

	for i, test := range tests {

		var eucl EuclideanRhythm

		err := eucl.Parse(test.start32th, test.n, test.m, test.dur)

		if err != nil {
			t.Errorf("[%v] error for Parse(%v,%v,%v,%v): %s", i, test.start32th, test.n, test.m, test.dur, err)
			continue
		}

		result := eucl.Sketch

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%v] Parse(%v,%v,%v,%v) = %#v \n                                               // expected: %#v", i, test.start32th, test.n, test.m, test.dur, result, test.expected)
		}
	}
}

func TestEuclideanRhythm(t *testing.T) {
	tests := []struct {
		n        int
		m        int
		expected []bool
	}{
		{3, 8, []bool{true, false, false, true, false, false, true, false}},

		// E(5,8) = [x . x x . x x .]
		{5, 8, []bool{true, false, true, true, false, true, true, false}},

		//	E(1,2) = [x .]
		{1, 2, []bool{true, false}},

		//	E(1,3) = [x . .]
		{1, 3, []bool{true, false, false}},

		// E(1,4) = [x . . .], etc.
		{1, 4, []bool{true, false, false, false}},

		//	E(4,12) = [x . . x . . x . . x . .],
		{4, 12, []bool{true, false, false, true, false, false, true, false, false, true, false, false}},

		// E(3,5)=[x . x . x]
		{3, 5, []bool{true, false, true, false, true}},

		// E(5,12) = [x . . x . x . . x . x .]
		{5, 12, []bool{true, false, false, true, false, true, false, false, true, false, true, false}},

		/*

			Note that since we are interested in cyclic non-periodic rhythms it is not necessary to enumerate these
			rhythms with multiples of k and n. For example, multiplying (1,3) by 4 gives (4,12) which yields:

			E(2,3) = [x . x]
			E(2,5)=[x . x . .]
			E(3,4)=[x . x x]

			E(3,5)=[x . x . x], when started on the second onset, is another thirteenth century Persian rhythm by the
			name of Khafif-e-ramal [34], as well as a Rumanian folk-dance rhythm [25].
			E(3,7)=[x . x . x . .] is a Ruchenitza rhythm used in a Bulgarian folk-dance [24]. It is also the metric
			pattern of Pink Floyd’s Money [17].
			E(3,8)=[x . . x . . x .] is the Cuban tresillo pattern discussed in the preceding [15].
			E(4,7)=[x . x . x . x] is another Ruchenitza Bulgarian folk-dance rhythm [24].
			E(4,9) = [x . x . x . x . .] is the Aksak rhythm of Turkey [6]. It is also the metric pattern used by Dave
			Brubeck in his piece Rondo a la Turk [17].
			E(4,11) = [x . . x . . x . . x .] is the metric pattern used by Frank Zappa in his piece titled Outside Now [17].
			E(5,6)=[x . x x x x] yields the York-Samai pattern, a popular Arab rhythm, when started on the second
			onset [30].
			E(5,7)=[x . x x . x x] is the Nawakhat pattern, another popular Arab rhythm [30].
			E(5,8)=[x . x x . x x .] is the Cuban cinquillo pattern discussed in the preceding [15]. When it is started
			on the second onset it is also the Spanish Tango [13] and a thirteenth century Persian rhythm, the Al-saghil-
			al-sani [34].
			E(5,9)=[x . x . x . x . x] is a popular Arab rhythm called Agsag-Samai [30]. When started on the second
			onset, it is a drum pattern used by the Venda in South Africa [26], as well as a Rumanian folk-dance
			rhythm [25].
			E(5,11)=[x . x . x . x . x . .] is the metric pattern used by Moussorgsky in Pictures at an Exhibition [17].
			E(5,12) = [x . . x . x . . x . x .] is the Venda clapping pattern of a South African children’s song [24].
			E(5,16) = [x . . x . . x . . x . . x . . . .] is the Bossa-Nova rhythm necklace of Brazil. The actual Bossa-
			Nova rhythm usually starts on the third onset as follows: [x . . x . . x . . . x . . x . .] [31]. However, there are
			other starting places as well, as for example [x . . x . . x . . x . . . x . .] [3].
			E(7,8) = [x . x x x x x x] i

			E(7,12) = [x . x x . x . x x . x .] is a common West African bell pattern. For example, it is used in the Mpre
			rhythm of the Ashanti people of Ghana [32].
			E(7,16) = [x . . x . x . x . . x . x . x .] is a Samba rhythm necklace from Brazil. The actual Samba rhythm
			is [x . x . . x . x . x . . x . x .] obtained by starting E(7,16) on the last onset. When E(7,16) is started on the
			fifth onset it is a clapping pattern from Ghana [24].
			E(9,16) = [x . x x . x . x . x x . x . x .] is a rhythm necklace used in the Central African Republic [2].
			When it is started on the fourth onset it is a rhythm played in West and Central Africa [15], as well as a
			cow-bell pattern in the Brazilian samba [29]. When it is started on the penultimate onset it is the bell pattern
			of the Ngbaka-Maibo rhythms of the Central African Republic [2].
			E(11,24) = [x . . x . x . x . x . x . . x . x . x . x . x .] is a rhythm necklace of the Aka Pygmies of Central
			Africa [2]. It is usually started on the seventh onset.
			E(13,24) = [x . x x . x . x . x . x . x x . x . x . x . x .] is another rhythm necklace of the Aka Pygmies of
			the upper Sangha [2]. It is usually started on the fourth onset.

			E(2,5)=[x . x . .] = (23) (classical, jazz, and Persian).
			E(3,7)=[x . x . x . .] = (223) (Bulgarian folk).
			E(4,9) = [x . x . x . x . .] = (2223) (Turkey).
			E(5,11)=[x . x . x . x . x . .] = (22223) (classical).
			E(5,16) = [x . . x . . x . . x . . x . . . .] = (33334) (Brazilian necklace).

			E(2,3) = [x . x] = (21) (West Africa, Latin America).
			E(3,4)=[x . x x] = (211) (Trinidad, Persia).
			E(3,5)=[x . x . x] = (221) (Rumanian and Persian necklaces).
			E(3,8)=[x . . x . . x .] = (332) (West Africa).
			E(4,7)=[x . x . x . x] = (2221) (Bulgaria).
			E(4,11) = [x . . x . . x . . x .] = (3332) (Frank Zappa).
			E(5,6)=[x . x x x x] = (21111) (Arab).
			E(5,7)=[x . x x . x x] = (21211) (Arab).
			E(5,9)=[x . x . x . x . x] = (22221) (Arab rhythm, South African and Rumanian necklaces).
			E(5,12) = [x . . x . x . . x . x .] = (32322) (South Africa).
			E(7,8) = [x . x x x x x x] = (2111111) (Tuareg rhythm of Libya).
			E(7,16) = [x . . x . x . x . . x . x . x .] = (3223222) (Brazilian necklace).
			E(11,24) = [x . . x . x . x . x . x . . x . x . x . x . x .] = (32222322222) (Central Africa).

			E(5,8)=[x . x x . x x .] = (21212) (West Africa).
			E(7,12) = [x . x x . x . x x . x .] = (2122122) (West Africa).
			E(9,16) = [x . x x . x . x . x x . x . x .] = (212221222) (West and Central African, and Brazilian necklaces).
			E(13,24) = [x . x x . x . x . x . x . x x . x . x . x . x .] = (2122222122222) (Central African necklace).
		*/

		//             [1      0      0      1     0      1      0     0     1       0    1      0     0]
		{5, 13, []bool{true, false, false, true, false, true, false, false, true, false, true, false, false}},
	}

	for i, test := range tests {
		result := euclideanRhythm(test.n, test.m)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%v] EuclideanRhythm(%v,%v) = %v // expected: %v", i, test.n, test.m, result, test.expected)
		}
	}
}

func TestHalfTonesToPitchbend(t *testing.T) {
	tests := []struct {
		halftones float64
		expected  int16
	}{
		{0, 0},
		{1, 4096},
		{2, 8192},
		{-2, -8191},
		{-1, -4096},
	}

	for i, test := range tests {
		got := HalfTonesToPitchbend(test.halftones, 2)

		if got != test.expected {
			t.Errorf("[%v] halfTonesToPitchbend(%v) == %v // expected %v", i, test.halftones, got, test.expected)
		}
	}
}

func TestPositionTo32th(t *testing.T) {
	tests := []struct {
		lastBeat  uint
		pos       string
		completed string
		num32th   uint
	}{
		{0, "1", "1", 0},
		{1, ";", "1;", 1},
		{1, ",", "1,", 2},
		{1, "&", "1&", 4},
		{1, "&;", "1&;", 5},
		//{0, ";", "1;", 1},
		{1, "&,", "1&,", 6},
		//{"1&", ".", "1&.", 6},
		{1, "&,;", "1&,;", 7},
		//{"1&", ".;", "1&.;", 7},
		//{"1&.", ";", "1&.;", 7},
		{1, "2", "2", 8},
		{2, "&", "2&", 12},
	}

	for i, test := range tests {
		completed, num32th, _ := PositionTo32th(test.lastBeat, test.pos)

		if completed != test.completed || num32th != test.num32th {
			t.Errorf("[%v] positionTo32th(%#v, %#v) = %#v, %v // expected %#v, %v", i, test.lastBeat, test.pos, completed, num32th, test.completed, test.num32th)
		}
	}
}

func TestGetQNNumberFromPos(t *testing.T) {
	tests := []struct {
		position string
		qnumber  uint
		rest     string
	}{
		{"1", 1, ""},
		{"2", 2, ""},
		{"3", 3, ""},
		{"4", 4, ""},
		{"5", 5, ""},
		{"1&", 1, "&"},
		{"1&;", 1, "&;"},
		{"1&,;", 1, "&,;"},
		{"1,;", 1, ",;"},
		{"1,", 1, ","},
		{"1&,", 1, "&,"},
		{"1;", 1, ";"},
		{"2&", 2, "&"},
		{"3&", 3, "&"},
		{"4&", 4, "&"},
		{"5&", 5, "&"},
	}

	for i, test := range tests {
		qn, rest := GetQNNumberFromPos(test.position)

		if qn != test.qnumber || rest != test.rest {
			t.Errorf("[%v] getQNNumberFromPos(%#v) = %v, %#v // expected %v, %#v", i, test.position, qn, rest, test.qnumber, test.rest)
		}
	}
}

func TestVelocityFromDynamic(t *testing.T) {
	tests := []struct {
		dyn string
		vel int8
	}{
		{"++++", 123},
		{"+++", 108},
		{"++", 93},
		{"+", 78},
		{"", 63},
		{"=", -1},
		{"-", 48},
		{"--", 33},
		{"---", 18},
		{"----", 5},
	}

	for _, test := range tests {
		got := DynamicToVelocity(test.dyn, 1, 127, 4, 15, 63)

		if got != test.vel {
			t.Errorf("velocityFromDynamic(%#v) = %v // expected %v", test.dyn, got, test.vel)
		}
	}
}

func TestPos32thToString(t *testing.T) {

	tests := []struct {
		pos      uint
		expected string
	}{
		{0, "1"},
		{1, "1;"},
		{2, "1,"},
		{3, "1,;"},
		{4, "1&"},
		{5, "1&;"},
		{6, "1&,"},
		{7, "1&,;"},
		{8, "2"},
		{9, "2;"},
		{10, "2,"},
		{11, "2,;"},
		{12, "2&"},
		{13, "2&;"},
		{14, "2&,"},
		{15, "2&,;"},
		{16, "3"},
		{24, "4"},
		{32, "5"},
		{40, "6"},
	}

	for _, test := range tests {
		got := Pos32thToString(test.pos)

		if got != test.expected {
			t.Errorf("pos32thToString(%v) = %#v // expected %#v", test.pos, got, test.expected)
		}
	}
}

func TestCalc32thsAdd(t *testing.T) {
	tests := []struct {
		distance32th int64
		notediff     int64
		expected     float64
	}{
		{32, 32, 1},
		{32, -32, -1},
		{32, 64, 2},
		{32, -64, -2},
	}

	for i, test := range tests {
		got := calcAdd(test.distance32th, float64(test.notediff))

		if got != test.expected {
			t.Errorf("[%v] calc32thsAdd(%v, %v) == %v // expected %v", i, test.distance32th, test.notediff, got, test.expected)
		}
	}
}
