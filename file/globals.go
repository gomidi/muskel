package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//var regAllDashes = regexp.MustCompile("^[-]+$")

func newBlankFile() *File {
	return &File{
		multiComments: map[int]*MultiLineComment{},
		commentLines:  map[int]string{},
	}
}

func FromReader(rd io.Reader, s score) (f *File) {
	f = newBlankFile()
	f.Score = s
	f.Input = bufio.NewScanner(rd)
	return f
}

func New(fpath string, s score) (f *File, err error) {
	fpath, err = filepath.Abs(fpath)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("%q is a directory", fpath)
	}

	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	f = FromReader(file, s)
	f.file = file
	f.dir = filepath.Dir(fpath)
	f.name = filepath.Base(fpath)
	return f, nil
}
