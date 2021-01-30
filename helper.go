package muskel

import (
	"fmt"
	"path/filepath"

	"gitlab.com/gomidi/muskel/score"
)

func init() {
	score.FindInclude = findInclude
}

var FILE_EXTENSION = ".mskl"

func findInclude(relDir string, file string) (resolved string, err error) {
	switch filepath.Ext(file) {
	case FILE_EXTENSION, ".md":
	default:
		file = file + FILE_EXTENSION
	}

	if filepath.IsAbs(file) {
		if score.FileExists(file) {
			//fmt.Printf("found: %q\n", file)
			return filepath.Abs(file)
		}
		return "", fmt.Errorf("file not found: %q (abs path)", file)
	}

	try0 := filepath.Join(relDir, file)
	if score.FileExists(try0) {
		//fmt.Printf("found: %q\n", try2)
		return filepath.Abs(try0)
	}

	try1 := filepath.Join(WORKING_DIR, file)
	if score.FileExists(try1) {
		//fmt.Printf("found: %q\n", try1)
		return filepath.Abs(try1)
	}

	try2 := filepath.Join(USER_DIR, file)
	if score.FileExists(try2) {
		//fmt.Printf("found: %q\n", try2)
		return filepath.Abs(try2)
	}

	return "", fmt.Errorf("file not found: %q", file)
}
