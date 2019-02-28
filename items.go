package muskel

// TODO see the following list

/*
+ patterns and definition
+ parts und sprünge
+ Glissandi!
+ accel. decel. cresc. descr. (wahrscheinlich mit Volume CC)
+ MIDI controls
- temperaments
- ggf. Chords
- translate augmentation symbols like ^ and ° to pitchbend messages
*/

type Note struct {
	letter    string
	octave    int
	augmenter string
	//dynamic        string
	velocity       int8
	glissandoStart bool
	dotted         bool
}

func (n Note) Dup() Note {
	return n
}

type DrumNote struct {
	name    string
	dynamic string
}

type MIDINote struct {
	note     int8
	velocity int8
	dotted   bool
}

func (mn MIDINote) Dup() MIDINote {
	return mn
}

type MIDICC [2]uint8
type MIDIPitchbend int16
type MIDIPolyAftertouch [2]uint8
type MIDIAftertouch uint8

type OSCMessage struct {
	Path string
	Args []interface{}
}

type NTuple struct {
	endPos uint // 32th in bar from the beginning
	items  []interface{}
}
type RepeatLastEvent struct{}
type RepeatLastBar struct{}
type RepeatLastBarUntilChange struct{}
type RepeatLastNBarsUntilChange int
type rest struct{}
type hold struct{}
type Lyric string

type RandomProbability struct {
	prob             uint8
	item             interface{}
	chosen           bool
	itemOriginalData string
}

type RandomChooser struct {
	alternatives             []interface{}
	alternativesOriginalData []string
	chosen                   int
}

var Hold = hold{}
var Rest = rest{}

/*
- notes, e.g. a'
- drumnotes e.g. sn
- motivs e.g. $a
- MIDI controller events e.g. PB200
- rests e.g. _
- samples with a timespan e.g. Sa
- samples with no timespan ("oneshot") e.g. Za
- repitition of last event e.g. %
- multiitems, consisting of several items happening at the same time. e.g. (a',PB200)
*/

type MultiItem []interface{}
