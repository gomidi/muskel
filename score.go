package muskel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
)

/*
TODO create an unrolled score and make conversion between the two.
SMF may only be written to an unrolled score
*/

type Score struct {
	Instruments []*Instrument
	Bars        []*Bar

	Meta               map[string]string
	Temperament        map[string]string
	PatternDefinitions map[string]*PatternDefinition
	Parts              map[string][2]int // part name to start bar no and last bar
	Comments           map[int]string    // line calculated from the start of the system
	HeaderComments     []string          // comments in the header, as they come

	isUnrolled        bool
	barNumbersTracked bool
}

func NewScore() *Score {
	return &Score{
		Meta:               map[string]string{},
		Temperament:        map[string]string{},
		PatternDefinitions: map[string]*PatternDefinition{},
		Parts:              map[string][2]int{},
		Comments:           map[int]string{},
		isUnrolled:         false,
	}
}

func (s *Score) Unroll() (*Score, error) {

	for _, b := range s.Bars {
		b.positions = make([]uint8, len(b.originalPositions))
		var oldPos string

		for pi, pos := range b.originalPositions {
			oldPos, b.positions[pi] = positionTo32th(oldPos, pos)
		}
	}

	nu := NewScore()
	nu.isUnrolled = true
	nu.Meta = s.Meta
	nu.Comments = s.Comments
	nu.Parts = s.Parts
	nu.Temperament = s.Temperament
	nu.PatternDefinitions = s.PatternDefinitions
	nu.HeaderComments = s.HeaderComments
	nu.barNumbersTracked = true
	err := s.unrollInstruments(nu)
	return nu, err
}

func (s *Score) WriteSMF(midifile string) error {

	if !s.isUnrolled {
		ur, err := s.Unroll()
		if err != nil {
			return err
		}
		return ur.WriteSMF(midifile)
	}

	numTracks := uint16(1) // first track is for tempo and time signatures

	for _, instr := range s.Instruments {
		if instr.MIDIChannel >= 0 {
			numTracks++
		}
	}

	sw := NewSMFWriter(s)

	return mid.NewSMFFile(midifile, numTracks, sw.Write,
		smfwriter.TimeFormat(smf.MetricTicks(960)),
		//		smfwriter.Debug(log.New(os.Stdout, "write MIDI", log.Lshortfile)),
	)
}

func (s *Score) WriteTo(wr io.Writer) (err error) {
	/*
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("Error: %v", r)
			}
		}()
	*/

	var bf bytes.Buffer

	fm := NewFormatter(s)
	fm.Format(&bf)

	_, err = bf.WriteTo(wr)
	return
}

func (s *Score) WriteToFile(filepath string) (err error) {
	dir, err := ioutil.TempDir(".", "muskel-fmt")
	if err != nil {
		return err
	}

	base := path.Base(filepath)

	f, err := os.Create(path.Join(dir, base))
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
		f.Close()
		os.RemoveAll(dir)
	}()

	err = s.WriteTo(f)
	if err != nil {
		return
	}

	err = os.Rename(path.Join(dir, base), filepath)
	return
}
