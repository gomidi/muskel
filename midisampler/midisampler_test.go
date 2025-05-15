package midisampler

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/v2/smf"
)

var testStr = `
{
	"format": 1,
	"timeformat": {
		"metricticks": 480
	},
	"tracks": [
		[
			{
				"data": {
					"channel": 3,
					"key": {{ .key }},
					"keyname": "C1",
					"velocity": 120
				},
				"delta": 24,
				"type": "noteon"
			},
			{{ zero 12 }}
			{
				"data": {
					"channel": 3,
					"key": {{ .key }},
					"keyname": "C1"
				},
				"delta": 24,
				"type": "noteoff"
			}
		]
	]
}
`

var expected = strings.TrimSpace(strings.ReplaceAll(`
{
	"format": 0,
	"timeformat": {
		"metricticks": 960
	},
	"tracks": [
		[
			{
				"data": {
					"channel": 3,
					"key": 65,
					"keyname": "F5",
					"velocity": 120
				},
				"delta": 48,
				"type": "noteon"
			},
			{
				"data": {
					"channel": 3,
					"key": 65,
					"keyname": "F5"
				},
				"delta": 72,
				"type": "noteoff"
			}
		]
	]
}
`, "\t", "  "))

var expectedAutoStop = strings.TrimSpace(strings.ReplaceAll(`
{
	"format": 0,
	"timeformat": {
		"metricticks": 960
	},
	"tracks": [
		[
			{
				"data": {
					"channel": 3,
					"key": 65,
					"keyname": "F5",
					"velocity": 120
				},
				"delta": 48,
				"type": "noteon"
			},
			{
				"data": "autostop",
				"delta": 40,
				"type": "text"
			},
			{
				"data": {
					"channel": 3,
					"key": 65,
					"keyname": "F5"
				},
				"delta": 0,
				"type": "noteoff"
			}
		]
	]
}
`, "\t", "  "))

func TestRunJSONTemplate(t *testing.T) {
	tr, offset, err := RunJSONTemplate(testStr, 960, 4000, map[string]any{"key": 65})

	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	var sm smf.SMF
	sm.TimeFormat = smf.MetricTicks(960)
	err = sm.Add(tr)

	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	bt, err := sm.MarshalJSONIndent()

	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	got := string(bt)

	if got != expected {
		t.Errorf("got:\n%s\nexpected:\n%s\n", got, expected)
	}

	if offset != 72 {
		t.Errorf("got offset %v // expected: %v", offset, 72)
	}
}

func TestRunJSONTemplateAutoStop(t *testing.T) {
	tr, offset, err := RunJSONTemplate(testStr, 960, 40, map[string]any{"key": 65})

	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	var sm smf.SMF
	sm.TimeFormat = smf.MetricTicks(960)
	err = sm.Add(tr)

	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	bt, err := sm.MarshalJSONIndent()

	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	got := string(bt)

	if got != expectedAutoStop {
		t.Errorf("got:\n%s\nexpected:\n%s\n", got, expectedAutoStop)
	}

	if offset != 72 {
		t.Errorf("got offset %v // expected: %v", offset, 72)
	}
}
