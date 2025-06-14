package smf

import (
	"fmt"
	"io"
	"strings"

	"gitlab.com/golang-utils/fs"

	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

const DEBUG_TEST = false

//var DEBUG bool = false

func WriteSMFTo(s *score.Score, wr io.Writer, filegroup string, opts ...Option) error {
	numTracks := uint16(1) // first track is for time signatures and tempo changes

	if items.DEBUG {
		fmt.Printf("WriteSMFTo filegroup %v\n", filegroup)
	}

	for trackName := range s.Unrolled {
		track, err := s.GetTrack(trackName)

		if err != nil {
			return fmt.Errorf("could not find track %q", trackName)
		}

		if track.MIDIChannel >= 0 && (filegroup == "*" || track.FileGroup == filegroup) {
			numTracks++
		}
	}

	wrt := NewWriter(wr, numTracks, 960, opts...)
	if items.DEBUG || DEBUG_TEST {
		//	wrt.SMF.Logger = debugLog{}
	}
	sw := New(s, filegroup, wrt)
	return sw.write()

}

func WriteFile(s *score.Score, midifile path.Local) (err error) {
	fsys, err := rootfs.New()
	if err != nil {
		return err
	}

	return WriteFileFS(s, fsys, midifile.RootRelative())
}

// pdffile is a string that is a relative path, but might have and %s placeholder
//func WriteFileFS(s *score.Score, fsys fs.FS, pdffile path.Relative) (err error) {

// WriteFile writes the score to the given SMF file
func WriteFileFS(s *score.Score, fsys fs.FS, midifile path.Relative) (err error) {
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

	hasPlaceholder := strings.Index(midifile.String(), "%s") > -1

	if !hasPlaceholder {
		return writeSMFToFile(s, fsys, midifile, "*")
	}

	var fileGroups = map[string]string{}
	for _, track := range s.Tracks {
		fileGroups[track.FileGroup] = fmt.Sprintf(midifile.String(), track.FileGroup)
	}

	if items.DEBUG {
		fmt.Printf("filegroups: %#v\n", fileGroups)
	}

	var errs errors

	for grp, fl := range fileGroups {
		err := writeSMFToFile(s, fsys, path.Relative(fl), grp)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
