package muskel

import (
	"fmt"

	"gitlab.com/golang-utils/fs"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel/score"
)

func init() {
	score.FindInclude = findInclude
}

var FILE_EXTENSION = ".mskl"

func _findInclude(fsys fs.FS, relDir path.Relative, file string) (resolved path.Relative, err error) {
	if fsys.Exists(path.Relative(file)) {
		return path.Relative(file), nil
	}

	try0 := relDir.Join(file)
	if fsys.Exists(try0) {
		return try0, nil
	}

	try1 := WORKING_DIR.RootRelative().Join(file)
	if fsys.Exists(try1) {
		return try1, nil
	}

	try2 := USER_DIR.RootRelative().Join(file)
	if fsys.Exists(try2) {
		return try2, nil
	}

	return "", fmt.Errorf("file not found: %q", file)
}

func findInclude(fsys fs.FS, relDir path.Relative, file string) (resolved path.Relative, err error) {
	if file == "" {
		panic("can't find empty file")
	}
	switch path.Ext(path.Relative(file)) {
	case FILE_EXTENSION, ".md", ".mskl", ".xlsx", ".csv":
		return _findInclude(fsys, relDir, file)
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
			resolved, err = _findInclude(fsys, relDir, file+ext)
			if err == nil {
				//fmt.Printf("found: %q\n", resolved)
				return
			}
		}
		return
	}
}
