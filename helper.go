package muskel

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"gitlab.com/gomidi/muskel/score"
)

func init() {
	score.FindInclude = findInclude
}

var FILE_EXTENSION = ".mskl"

func absNormalized(file string) (string, error) {
	s, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(s, "\\", "/"), nil
}

func _findInclude(relDir string, file string) (resolved string, err error) {
	if filepath.IsAbs(file) {
		if score.FileExists(file) {
			return absNormalized(file)
		}
		return "", fmt.Errorf("file not found: %q (abs path)", file)
	}

	relDir = strings.ReplaceAll(relDir, "\\", "/")
	file = strings.ReplaceAll(file, "\\", "/")

	try0 := path.Join(relDir, file)
	try0 = strings.ReplaceAll(try0, "\\", "/")
	if score.FileExists(try0) {
		//fmt.Printf("found: %q\n", try0)
		return absNormalized(try0)
	}

	try1 := path.Join(WORKING_DIR, file)
	try1 = strings.ReplaceAll(try1, "\\", "/")
	if score.FileExists(try1) {
		//fmt.Printf("found: %q\n", try1)
		return absNormalized(try1)
	}

	try2 := path.Join(USER_DIR, file)
	try2 = strings.ReplaceAll(try2, "\\", "/")
	if score.FileExists(try2) {
		//fmt.Printf("found: %q\n", try2)
		return absNormalized(try2)
	}

	return "", fmt.Errorf("file not found: %q", file)
}

func findInclude(relDir string, file string) (resolved string, err error) {
	if file == "" {
		panic("can't find empty file")
	}
	switch path.Ext(file) {
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
