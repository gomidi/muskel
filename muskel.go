package muskel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel/file"
	"gitlab.com/gomidi/muskel/image"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
	"gitlab.com/gomidi/muskel/smfimport"
	"gitlab.com/gomidi/muskel/xlsx"
	"gitlab.com/gomidi/smfimage"
	"gitlab.com/metakeule/config"
)

const MUSKEL_VERSION_FILE = "muskel_version.txt"

func ReadWDVersionFile(dir string) (*config.Version, error) {
	p := filepath.Join(dir, MUSKEL_VERSION_FILE)
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	v, err2 := config.ParseVersion(strings.TrimSpace(string(b)))
	if err2 != nil {
		return nil, err2
	}

	return v, nil
}

func WriteWDVersionFile(dir string) error {
	p := filepath.Join(dir, MUSKEL_VERSION_FILE)
	return ioutil.WriteFile(p, []byte(VERSION), 0644)
}

func Import(srcFile string, targetFile string, opts ...smfimport.Option) error {
	fh, err := os.Open(srcFile)

	if err != nil {
		return fmt.Errorf("can't open file %q: %s", srcFile, err.Error())
	}

	defer fh.Close()

	im := smfimport.New(srcFile, fh, opts...)

	if filepath.Ext(targetFile) == ".xlsx" {
		var tracksbf, scorebf bytes.Buffer
		err := im.WriteTracksAndScoreTable(&tracksbf, &scorebf)
		if err != nil {
			return fmt.Errorf("can't import from %q: %s", srcFile, err.Error())
		}
		return xlsx.Write(targetFile, tracksbf.String(), scorebf.String())
	} else {

		tg, err := os.Create(targetFile)

		if err != nil {
			return fmt.Errorf("can't create file %q: %s", targetFile, err.Error())
		}

		defer tg.Close()

		return im.WriteUnrolled(tg)
	}
}

func newFile(filename string, params []string, rd io.Reader, opts ...score.Option) *file.File {
	sc := score.New(filename, params, opts...)
	return file.FromReader(rd, sc)
}

func Format(filename string, params []string, rd io.Reader, wr io.Writer, opts ...score.Option) error {
	var f = newFile(filename, params, rd, opts...)
	err := f.Parse()

	if err != nil {
		return err
	}

	return f.WriteTo(wr)
}

func Unroll(mainFile string, params []string, rd io.Reader, wr io.Writer, opts ...score.Option) error {
	sc := score.New(mainFile, params, opts...)
	f := file.FromReader(rd, sc)

	err := f.Parse()
	if err != nil {
		return fmt.Errorf("parsing error: %s", err)
	}

	err = sc.Unroll()
	if err != nil {
		return err
	}

	sc.WriteUnrolled(wr)
	return nil
}

func WriteSMF(filename string, params []string, rd io.Reader, fmtwr io.Writer, midwr io.Writer, opts ...score.Option) error {
	var f = newFile(filename, params, rd, opts...)
	err := f.Parse()

	if err != nil {
		return err
	}

	if fmtwr != nil {
		err = f.WriteTo(fmtwr)

		if err != nil {
			return err
		}
	}

	sc := f.Score.(*score.Score)

	err = sc.Unroll()
	if err != nil {
		return err
	}

	return smf.WriteSMFTo(sc, midwr, "*")
}

func ParseFile(mainFile string, params []string, opts ...score.Option) (sc *score.Score, err error) {
	sc = score.New(mainFile, params, opts...)
	err = sc.Parse()
	return
}

func Convert(mainFile string, params []string, smffile string, opts ...score.Option) error {
	sc := score.New(mainFile, params, opts...)
	err := sc.Parse()
	if err != nil {
		return err
	}
	err = sc.Format()
	if err != nil {
		return err
	}

	err = sc.Unroll()
	if err != nil {
		return err
	}
	return smf.WriteFile(sc, smffile)
}

func WriteSMFFile(sc *score.Score, smffile string, opts ...smfwriter.Option) error {
	return smf.WriteFile(sc, smffile, opts...)
}

// WriteImage expects the given score to be unrolled
func WriteImage(sc *score.Score, outfile string, opts ...smfimage.Option) error {
	var midibf bytes.Buffer
	err := smf.WriteSMFTo(sc, &midibf, "*")

	if err != nil {
		return err
	}

	return image.MakeImage(sc, &midibf, outfile, opts...)
}
