package lilypond

import (
	"strings"
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
