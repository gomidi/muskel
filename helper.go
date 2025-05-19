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

	places := []path.Relative{
		// fs root
		path.Relative(file).Clean(),

		// related to current score dir
		relDir.Join(file).Clean(),

		// related to working directory (=project directory)
		WORKING_DIR.RootRelative().Join(file),

		// related to user directory
		USER_DIR.RootRelative().Join(file),
	}

	for _, fl := range places {
		if fsys.Exists(fl) {
			return fl, nil
		}
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
