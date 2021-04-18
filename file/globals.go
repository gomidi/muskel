package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

func NewCSV(fpath string, seperator rune, s score) (f *File, err error) {
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

	defer file.Close()

	muskel_string, err := csv.Read(file, seperator)
	f = FromReader(strings.NewReader(muskel_string), s)
	f.SetFileInfos(nil, fpath)
	return f, nil
}

func NewXLSX(fpath string, s score) (f *File, err error) {
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

	/*
		file, err := os.Open(fpath)
		if err != nil {
			return nil, err
		}

		defer file.Close()
	*/

	muskel_string, err := xlsx.Read(fpath)
	f = FromReader(strings.NewReader(muskel_string), s)
	f.SetFileInfos(nil, fpath)
	return f, nil
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

	defer file.Close()

	bt, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	f = FromReader(bytes.NewReader(bt), s)
	f.SetFileInfos(bt, fpath)
	return f, nil
}
