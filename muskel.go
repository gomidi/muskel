package muskel

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel/formatter"
	"gitlab.com/gomidi/muskel/parser"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
	"gitlab.com/gomidi/muskel/template"
)

// ParseFile parses the given file and returns the resulting Score
func ParseFile(filepath string) (sc *score.Score, err error) {
	return parser.ParseFile(filepath)
}

// Parse reads from the given reader and returns the resulting Score
func Parse(rd io.Reader, filepath string) (sc *score.Score, err error) {
	return parser.Parse(rd, filepath)
}

// Format writes the formatted score to the given writer (in a formatted way)
func WriteFormattedTo(s *score.Score, wr io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	formatter.New(s).WriteTo(wr)

	return nil
}

func WriteSMFFile(s *score.Score, midifile string, options ...smfwriter.Option) (err error) {
	return smf.WriteFile(s, midifile, options...)
}

// WriteFormattedFile writes the score to the given file (in a formatted way)
// It only writes to the file if the formatting was successful
func WriteFormattedFile(s *score.Score, filep string) (err error) {

	// dir, err := ioutil.TempDir(".", "muskel-fmt")
	dir, err := tempDir("muskel-fmt")
	if err != nil {
		return fmt.Errorf("can't create tempdir: %v", err)
	}

	base := path.Base(filep)

	f, err := os.Create(path.Join(dir, base))
	if err != nil {
		return fmt.Errorf("can't create tempdir: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			f.Close()
			err = fmt.Errorf("%v", r)
		}
		os.RemoveAll(dir)
	}()

	err = WriteFormattedTo(s, f)
	if err != nil {
		return
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("can't close file %q: %v", path.Join(dir, base), err)
	}

	err = os.Rename(path.Join(dir, base), filep)
	if err != nil {
		err = fmt.Errorf("can't move %q to %q", path.Join(dir, base), filep)
	}
	return
}

func RenameInstrument(s *score.Score, old, nu string) error {
	in := s.GetInstrument(old)
	if in == nil {
		return fmt.Errorf("instrument with name %q could not be found", old)
	}

	in.Name = nu
	err := WriteFormattedFile(s, s.FileName)

	var includes = map[string]bool{}
	for _, b := range s.Bars {
		if b.Include != "" {

			if includes[b.Include] {
				continue
			}

			target, err := parser.IncludeScore(s, b.Include)
			if err != nil {
				return err
			}

			includes[b.Include] = true

			RenameInstrument(target, old, nu)
		}
	}
	return err
}

func syncInclude(s *score.Score, include string) error {

	target, err := parser.IncludeScore(s, include)
	if err != nil {
		// fmt.Printf("ERROR while including %q\n", include)
		return err
	}

	// fmt.Printf("including %q\n", include)

	var changed bool

	for _, instr := range s.Instruments {
		if !target.HasInstrument(instr.Name) {
			target.AddInstrument(instr)
			changed = true
		} else {
			ti := target.GetInstrument(instr.Name)

			if ti == nil {
				return fmt.Errorf("could not get instrument %q from include %q although it says, is has this instrument", instr.Name, include)
			}

			ti.MIDIBank = instr.MIDIBank
			ti.MIDIChannel = instr.MIDIChannel
			ti.MIDIProgram = instr.MIDIProgram
			ti.MIDITranspose = instr.MIDITranspose
			ti.MIDIVolume = instr.MIDIVolume
			ti.FileGroup = instr.FileGroup
			ti.PitchbendRange = instr.PitchbendRange
			changed = true
		}
	}

	if changed {
		// fmt.Printf("writing include %q to %q\n", include, target.FileName)
		WriteFormattedFile(target, target.FileName)
	}

	return SyncInstruments(target)
}

func SyncInstruments(s *score.Score) error {
	var includes = map[string]bool{}
	for _, b := range s.Bars {
		if b.Include != "" {

			if includes[b.Include] {
				continue
			}

			includes[b.Include] = true

			err := syncInclude(s, b.Include)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func renameTemplateDefinition(s *score.Score, old, nu string) error {
	_, hasOwn := s.TemplateDefinitions[old]
	_, hasIncluded := s.IncludedTemplateDefinitions[old]

	if !s.HasTemplate(old) {
		return fmt.Errorf("could not find a template with the name %q", old)
	}

	if s.HasTemplate(nu) {
		return fmt.Errorf("a template with name %q already exists", nu)
	}

	if hasOwn {
		pt := s.TemplateDefinitions[old]

		delete(s.TemplateDefinitions, old)

		pt.Name = nu

		s.TemplateDefinitions[nu] = pt
		return renameTemplateInCalls(s, old, nu)
	}

	if hasIncluded {
		pt := s.IncludedTemplateDefinitions[old]

		delete(s.IncludedTemplateDefinitions, old)

		pt.Name = nu

		s.IncludedTemplateDefinitions[nu] = pt
	}

	return nil
}

func RenameTemplate(s *score.Score, old, nu string) (err error) {
	err = renameTemplateDefinition(s, old, nu)

	if err != nil {
		return
	}

	var includes = map[string]bool{}

	for _, b := range s.Bars {
		if b.Include != "" {

			if includes[b.Include] {
				continue
			}

			target, err := parser.IncludeScore(s, b.Include)
			if err != nil {
				return err
			}

			RenameTemplate(target, old, nu)
			//target.WriteToFile(target.FileName)

			includes[b.Include] = true
		}
	}

	renameTemplateInCalls(s, old, nu)

	for _, templ := range s.TemplateDefinitions {
		it := strings.Split(templ.Original, " ")

		var changed bool = false

		for i, thing := range it {
			stripped := strings.TrimSpace(thing)
			stripped = strings.TrimRight(stripped, ":")
			stripped = strings.TrimRight(stripped, "+")
			stripped = strings.TrimRight(stripped, "+")
			stripped = strings.TrimRight(stripped, "+")
			stripped = strings.TrimRight(stripped, "-")
			stripped = strings.TrimRight(stripped, "-")
			stripped = strings.TrimRight(stripped, "-")
			stripped = strings.TrimRight(stripped, "*")
			stripped = strings.TrimLeft(stripped, "1234567890")
			stripped = strings.TrimLeft(stripped, "1234567890")
			stripped = strings.TrimLeft(stripped, "!")

			if stripped == old {
				it[i] = strings.Replace(thing, old, nu, 1)
				changed = true
			}
		}

		if changed {
			templ.Original = strings.Join(it, " ")
		}
	}

	return WriteFormattedFile(s, s.FileName)

}

func tempDir(prefix string) (dir string, err error) {
	dir = fmt.Sprintf("%s-%v", prefix, time.Now().Unix())
	err = os.Mkdir(dir, 0775)
	return
}

func renameTemplateInCalls(s *score.Score, old, nu string) (err error) {
	for _, instr := range s.Instruments {
		for ii, bev := range instr.Events {
			for i, ev := range bev {
				switch v := ev.Item.(type) {
				case *template.Call:
					if v.Name == old {
						v.Name = nu
					}
					ev.Item = v
					ev.OriginalData = strings.Replace(ev.OriginalData, old, nu, 1)
					bev[i] = ev
				}
			}

			instr.Events[ii] = bev
		}
	}

	return nil

}
