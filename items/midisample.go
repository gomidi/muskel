package items

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel/midisampler"
)

var MIDI_SAMPLES_DIR = path.Relative("midisamples/")

var _ Item = &MidiSample{}

type MidiSample struct {
	FileName       string
	TrackNo        int
	Placeholders   map[string]any
	LengthTicks    uint
	NegativeOffset uint64
	Track          smf.Track
	posIn32th      uint
}

func (m MidiSample) Dup() Item {
	return &m
}

func (m *MidiSample) Fetch(fsys fs.ReadOnly, dir path.Relative, ticksPerQuarterNote uint16) error {
	if m.Placeholders == nil {
		m.Placeholders = map[string]any{}
	}

	bt, err := rootfs.ReadFile(path.MustWD().Join(dir.String(), m.FileName+".json"))

	//bt, err := fs.ReadFile(fsys, dir.Join(m.FileName+".json"))
	if err != nil {
		return err
	}

	m.Track, m.NegativeOffset, err = midisampler.RunJSONTemplate(string(bt), ticksPerQuarterNote, m.TrackNo, uint64(m.LengthTicks), m.Placeholders)
	return err
}

func (m *MidiSample) Parse(data string, posIn32th uint) (err error) {
	m.Placeholders = map[string]any{}
	m.posIn32th = posIn32th

	if len(data) < 2 || data[0] != '>' {
		return fmt.Errorf("invalid syntax for MidiSample: %q", data)
	}

	data = data[1:]

	var fname string

	if idx := strings.Index(data, "("); idx > 0 {
		fname = data[:idx]
		lidx := strings.LastIndex(data, ")")
		params := strings.TrimSpace(data[idx+1 : lidx])
		ps := splitParams(params)
		for _, param := range ps {
			iidx := strings.Index(param, "=")
			if iidx > 0 {
				m.Placeholders[params[:iidx]] = params[iidx:]
			}
		}
	} else {
		fname = data
	}

	fidx := strings.Index(data, "#")

	if fidx <= 0 {
		return fmt.Errorf("no tracknumber defined for midisample %q", fname)
	}

	m.FileName = fname[:fidx]
	no := fname[fidx+1:]

	n, err := strconv.Atoi(no)
	if err != nil {
		return fmt.Errorf("not a number %q in midisample %q", no, fname)
	}

	m.TrackNo = n
	return nil
}

func (m MidiSample) String() string {
	var bf bytes.Buffer

	for k, v := range m.Placeholders {
		fmt.Fprintf(&bf, "%s=%s,", k, v)
	}

	params := bf.String()
	if len(params) > 0 {
		params = params[:len(params)-1] // remove the last comma
	}

	if len(params) > 0 {
		return fmt.Sprintf("%s(%s)", m.FileName, params)
	}

	return m.FileName
}
