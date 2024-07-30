package smf

import (
	"fmt"
	"io"
	"strings"

	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

const DEBUG_TEST = false

//var DEBUG bool = false

// func WriteSMFTo(s *score.Score, wr io.Writer, filegroup string, options ...smfwriter.Option) error {
func WriteSMFTo(s *score.Score, wr io.Writer, filegroup string) error {
	numTracks := uint16(1) // first track is for time signatures and tempo changes

	if items.DEBUG {
		fmt.Printf("WriteSMFTo filegroup %v\n", filegroup)
	}

	for trackName := range s.Unrolled {

		//for _, track := range s.Tracks {
		//for name, col := range s.Columns {
		track, err := s.GetTrack(trackName)

		if err != nil {
			return fmt.Errorf("could not find track %q", trackName)
		}
		// fmt.Printf("track %q midichannel: %v\n", track.Name, track.MIDIChannel)
		if track.MIDIChannel >= 0 && (filegroup == "*" || track.FileGroup == filegroup) {
			numTracks++
		}
	}

	/*
		if DEBUG_TEST {
			options = append(options, smfwriter.Debug(debugLog{}))
		}
	*/

	//sw := newSMF(s, filegroup, newWriter(wr, numTracks, 960, options...))
	wrt := NewWriter(wr, numTracks, 960)
	if items.DEBUG || DEBUG_TEST {
		wrt.SMF.Logger = debugLog{}
	}
	sw := New(s, filegroup, wrt)
	return sw.write()

}

// WriteFile writes the score to the given SMF file
// func WriteFile(s *score.Score, midifile string, options ...smfwriter.Option) (err error) {
func WriteFile(s *score.Score, midifile string) (err error) {
	if items.DEBUG {
		fmt.Printf("WriteFile(%q) called\n", midifile)
	}
	defer func() {
		if !items.DEBUG {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	hasPlaceholder := strings.Index(midifile, "%s") > -1

	if !hasPlaceholder {
		return writeSMFToFile(s, midifile, "*")
	}

	var fileGroups = map[string]string{}
	for _, track := range s.Tracks {
		fileGroups[track.FileGroup] = fmt.Sprintf(midifile, track.FileGroup)
	}

	if items.DEBUG {
		fmt.Printf("filegroups: %#v\n", fileGroups)
	}

	var errs errors

	for grp, fl := range fileGroups {
		err := writeSMFToFile(s, fl, grp)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
