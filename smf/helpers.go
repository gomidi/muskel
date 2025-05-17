package smf

import (
	"fmt"
	"os"
	"strings"

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
func writeSMFToFile(s *score.Score, midifile, filegroup string, opts ...Option) error {
	f, err := os.Create(midifile)
	if err != nil {
		return err
	}

	if items.DEBUG {
		fmt.Printf("writing to %q\n", midifile)
	}

	defer f.Close()
	return WriteSMFTo(s, f, filegroup, opts...)
}
