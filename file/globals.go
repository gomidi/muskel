package file

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel/csv"
	"gitlab.com/gomidi/muskel/xlsx"
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

func NewCSV(fsys fs.FS, file path.Relative, seperator rune, s score) (f *File, err error) {
	if path.IsDir(file) {
		return nil, fs.ErrExpectedFile.Params(file.String())
	}

	fl, err := fs.OpenReader(fsys, file)

	if err != nil {
		return nil, err
	}

	muskel_string, err := csv.Read(fl, seperator)
	f = FromReader(strings.NewReader(muskel_string), s)
	f.SetFileInfos(nil, file)
	err = fl.Close()
	return f, err
}

func NewXLSX(fsys fs.FS, file path.Relative, s score) (f *File, err error) {
	if path.IsDir(file) {
		return nil, fs.ErrExpectedFile.Params(file.String())
	}

	fl, err := fs.OpenReader(fsys, file)

	if err != nil {
		return nil, err
	}

	muskel_string, err := xlsx.ReadFrom(fl)
	f = FromReader(strings.NewReader(muskel_string), s)
	f.SetFileInfos(nil, file)
	err = fl.Close()
	return f, err
}

func New(fsys fs.FS, file path.Relative, s score) (f *File, err error) {
	if path.IsDir(file) {
		return nil, fs.ErrExpectedFile.Params(file.String())
	}

	fl, err := fs.OpenReader(fsys, file)

	if err != nil {
		return nil, err
	}

	bt, err := io.ReadAll(fl)
	if err != nil {
		return nil, err
	}

	f = FromReader(bytes.NewReader(bt), s)
	f.SetFileInfos(bt, file)
	err = fl.Close()
	return f, err
}
