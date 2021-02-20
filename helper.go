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

func _findInclude(relDir string, file string) (resolved string, err error) {
	//fmt.Printf("try to find include %q in %q\n", file, relDir)
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

func findInclude(relDir string, file string) (resolved string, err error) {
	switch filepath.Ext(file) {
	case FILE_EXTENSION, ".md", ".mskl", ".xlsx", ".csv":
		return _findInclude(relDir, file)
	default:
		fallbacks := []string{FILE_EXTENSION}
		switch FILE_EXTENSION {
		case ".md":
			fallbacks = append(fallbacks, ".mskl")
		case ".mskl":
			fallbacks = append(fallbacks, ".md")
		case ".csv", ".xlsx":
			fallbacks = append(fallbacks, ".md")
			fallbacks = append(fallbacks, ".mskl")
		}

		for _, ext := range fallbacks {
			resolved, err = _findInclude(relDir, file+ext)
			if err == nil {
				//fmt.Printf("found: %q\n", resolved)
				return
			}
		}
		return
	}
}
