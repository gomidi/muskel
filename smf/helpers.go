package smf

import (
	"fmt"
	"os"
	"strings"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel/items"
	"gitlab.com/gomidi/muskel/score"
)

type errors []error

func (e errors) Error() string {
	var bf strings.Builder
	bf.WriteString("The following errors happened:\n")
	for _, err := range e {
		bf.WriteString(err.Error() + "\n")
	}

	return bf.String()
}

type debugLog struct{}

func (debugLog) Printf(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", vals...)
}

// func writeSMFToFile(s *score.Score, midifile, filegroup string, options ...smfwriter.Option) error {
func writeSMFToFile(s *score.Score, fsys fs.FS, midifile path.Relative, filegroup string, opts ...Option) error {
	f, err := fs.OpenWriter(fsys, midifile)

	if err != nil {
		return err
	}

	if items.DEBUG {
		fmt.Printf("writing to %q\n", midifile.ToSystem())
	}

	err = WriteSMFTo(s, f, filegroup, opts...)
	if err != nil {
		f.Close()
		return err
	}

	return f.Close()
}
