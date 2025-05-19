package lilypond

import (
	"fmt"
	"strings"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

func Create(s *score.Score, filegroup string, opts ...Option) (string, error) {
	sw := newScore(s, filegroup)
	return sw.String()
}

func CreateBook(s *score.Score, filegroup string, opts ...Option) (*lilypond.Book, error) {
	sw := newScore(s, filegroup)
	return sw.makeBook()
}

func WritePDFTo(s *score.Score, pdffile path.Local, filegroup string, opts ...Option) error {

	if items.DEBUG {
		fmt.Printf("writing to %q\n", pdffile.ToSystem())
	}

	if items.DEBUG {
		fmt.Printf("WritePDFTo filegroup %v\n", filegroup)
	}

	sw := newScore(s, filegroup)
	return sw.WritePDF(pdffile)
}

// pdffile is a string that is a relative path, but might have and %s placeholder
func WriteFile(s *score.Score, pdffile path.Local) (err error) {
	if items.DEBUG {
		fmt.Printf("WriteFile(%q) called\n", pdffile.String())
	}
	defer func() {
		if !items.DEBUG {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	hasPlaceholder := strings.Index(pdffile.String(), "%s") > -1

	if !hasPlaceholder {
		return WritePDFTo(s, pdffile, "*")
	}

	var fileGroups = map[string]string{}
	for _, track := range s.Tracks {
		fileGroups[track.FileGroup] = fmt.Sprintf(pdffile.String(), track.FileGroup)
	}

	if items.DEBUG {
		fmt.Printf("filegroups: %#v\n", fileGroups)
	}

	var errs errors

	for grp, fl := range fileGroups {
		err := WritePDFTo(s, path.MustLocal(fl), grp)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
