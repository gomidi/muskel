package smf

import (
	"fmt"
	"os"
	"strings"

	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/sketch"
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

func writeSMFToFile(s *score.Score, midifile, filegroup string, options ...smfwriter.Option) error {
	f, err := os.Create(midifile)
	if err != nil {
		return err
	}

	if sketch.DEBUG {
		fmt.Printf("writing to %q\n", midifile)
	}

	defer f.Close()
	return WriteSMFTo(s, f, filegroup)
}
