package stops_handling

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/v2"
)

func TestOverride(t *testing.T) {

	tests := []struct {
		posItems PositionedItems
		expected string
	}{
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    NoteItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
					},
				},
				{
					pos32th: 16,
					Item:    NoteItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1920] MetaSeqData bytes: 7F 7F
[1920] NoteOn channel: 2 key: 26 velocity: 90
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    NoteItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
					},
				},
				{
					pos32th: 16,
					Item:    NoteOnItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1920] NoteOn channel: 2 key: 26 velocity: 90
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    NoteOnItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
					},
				},
				{
					pos32th: 16,
					Item:    NoteItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
				{
					pos32th: 32,
					Item:    NoteOffItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOff(2, midi.C(2)).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1920] NoteOn channel: 2 key: 26 velocity: 90
[3840] NoteOff channel: 2 key: 24
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
				{
					pos32th: 32,
					Item:    NoteOnItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.G(2), 110).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1920] MetaSeqData bytes: 7F 7F
[1920] NoteOn channel: 2 key: 26 velocity: 90
[3840] MetaSeqData bytes: 7F 7F
[3840] NoteOn channel: 2 key: 31 velocity: 110
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
				{
					pos32th: 32,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.G(2), 110).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.A(2), 110).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1920] MetaSeqData bytes: 7F 7F
[1920] NoteOn channel: 2 key: 26 velocity: 90
[3840] MetaSeqData bytes: 7F 7F
[3840] NoteOn channel: 2 key: 31 velocity: 110
[5760] MetaSeqData bytes: 7F 7F
[5760] NoteOn channel: 2 key: 33 velocity: 110
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    MidiSampleItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
						{pos32th: 12, msg: midi.NoteOff(2, midi.C(2)).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
						{pos32th: 64, msg: midi.NoteOff(2, midi.D(2)).Bytes()},
					},
				},
				{
					pos32th: 32,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.G(2), 110).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.A(2), 110).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1440] NoteOff channel: 2 key: 24
[1920] NoteOn channel: 2 key: 26 velocity: 90
[3840] MetaSeqData bytes: 7F 7F
[3840] NoteOn channel: 2 key: 31 velocity: 110
[5760] MetaSeqData bytes: 7F 7F
[5760] NoteOn channel: 2 key: 33 velocity: 110
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
				{
					pos32th: 32,
					Item:    MidiSampleItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
						{pos32th: 12, msg: midi.NoteOff(2, midi.C(2)).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
						{pos32th: 64, msg: midi.NoteOff(2, midi.D(2)).Bytes()},
					},
				},
				{
					pos32th: 128,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.G(2), 110).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.A(2), 110).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1920] MetaSeqData bytes: 7F 7F
[1920] NoteOn channel: 2 key: 26 velocity: 90
[3840] MetaSeqData bytes: 7F 7F
[3840] NoteOn channel: 2 key: 24 velocity: 100
[5280] NoteOff channel: 2 key: 24
[5760] NoteOn channel: 2 key: 26 velocity: 90
[11520] NoteOff channel: 2 key: 26
[15360] MetaSeqData bytes: 7F 7F
[15360] NoteOn channel: 2 key: 31 velocity: 110
[17280] MetaSeqData bytes: 7F 7F
[17280] NoteOn channel: 2 key: 33 velocity: 110
`,
		},
		{
			PositionedItems{
				{
					pos32th: 0,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(2), 100).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(2), 90).Bytes()},
					},
				},
				{
					pos32th: 32,
					Item:    MidiSampleItem,
					Offset:  -2000,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.C(3), 100).Bytes()},
						{pos32th: 12, msg: midi.NoteOff(2, midi.C(3)).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.D(3), 90).Bytes()},
						{pos32th: 64, msg: midi.NoteOff(2, midi.D(3)).Bytes()},
					},
				},
				{
					pos32th: 128,
					Item:    PatternItem,
					evts: Events{
						{pos32th: 0, msg: midi.NoteOn(2, midi.G(2), 110).Bytes()},
						{pos32th: 16, msg: midi.NoteOn(2, midi.A(2), 110).Bytes()},
					},
				},
			},
			`
[0] NoteOn channel: 2 key: 24 velocity: 100
[1840] MetaSeqData bytes: 7F 7F
[1840] NoteOn channel: 2 key: 36 velocity: 100
[3280] NoteOff channel: 2 key: 36
[3760] NoteOn channel: 2 key: 38 velocity: 90
[9520] NoteOff channel: 2 key: 38
[15360] MetaSeqData bytes: 7F 7F
[15360] NoteOn channel: 2 key: 31 velocity: 110
[17280] MetaSeqData bytes: 7F 7F
[17280] NoteOn channel: 2 key: 33 velocity: 110
`,
		},
	}

	for i, test := range tests {
		evts := GetTrackEvents(960, 200000, test.posItems)

		var bf bytes.Buffer

		for _, ev := range evts {
			fmt.Fprintf(&bf, "[%v] %s\n", ev.absPosition, ev.msg.String())
		}

		var expected = strings.TrimSpace(test.expected)

		got := strings.TrimSpace(bf.String())

		if got != expected {
			t.Errorf("[%v] got:\n%s\nexpected:\n%s\n", i, got, expected)
		}
	}

}
