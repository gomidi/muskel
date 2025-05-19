package muskel

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/filesystems/dirfs"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/golang-utils/version/v2"

	//	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/midi/tools/smfimage"
	"gitlab.com/gomidi/muskel/file"
	"gitlab.com/gomidi/muskel/image"
	"gitlab.com/gomidi/muskel/lilypond"
	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/muskel/smf"
)

const MUSKEL_VERSION_FILE = "muskel_version.txt"

func ReadWDVersionFile(dir path.Local) (*version.Version, error) {
	fsys, err := rootfs.New()

	if err != nil {
		return nil, err
	}

	p := dir.RootRelative().Join(MUSKEL_VERSION_FILE)
	b, err := fs.ReadFile(fsys, p)
	if err != nil {
		return nil, err
	}
	v, err2 := version.Parse(strings.TrimSpace(string(b)))
	if err2 != nil {
		return nil, err2
	}

	return v, nil
}

func WriteWDVersionFile(dir string) error {
	p := filepath.Join(dir, MUSKEL_VERSION_FILE)
	v := version.BuildVersion()
	return os.WriteFile(p, []byte(v.String()), 0644)
}

/*
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
*/

func newFile(filename path.Local, params []string, rd io.Reader, opts ...score.Option) *file.File {
	fsys, err := dirfs.New(filename.Dir())

	if err != nil {
		panic(err.Error())
	}

	opts = append(opts, score.FS(fsys))

	sc := score.New(path.Relative(path.Name(filename)), params, opts...)
	return file.FromReader(rd, sc)
}

func Format(filename path.Local, params []string, rd io.Reader, wr io.Writer, opts ...score.Option) error {
	var f = newFile(filename, params, rd, opts...)
	err := f.Parse()

	if err != nil {
		return err
	}

	return f.WriteTo(wr)
}

func FormatFS(fsys fs.FS, filename path.Relative, params []string, rd io.Reader, wr io.Writer, opts ...score.Option) error {
	opts = append(opts, score.FS(fsys))

	sc := score.New(path.Relative(path.Name(filename)), params, opts...)
	f := file.FromReader(rd, sc)

	err := f.Parse()

	if err != nil {
		return err
	}

	return f.WriteTo(wr)
}

func UnrollFS(fsys fs.FS, mainFile path.Relative, params []string, rd io.Reader, wr io.Writer, opts ...score.Option) error {
	opts = append(opts, score.FS(fsys))

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

func Unroll(mainFile path.Local, params []string, rd io.Reader, wr io.Writer, opts ...score.Option) error {
	fsys, err := dirfs.New(mainFile.Dir())

	if err != nil {
		panic(err.Error())
	}

	opts = append(opts, score.FS(fsys))

	sc := score.New(path.Relative(path.Name(mainFile)), params, opts...)
	f := file.FromReader(rd, sc)

	err = f.Parse()
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

func WriteSMF(filename path.Local, params []string, rd io.Reader, fmtwr io.Writer, midwr io.Writer, opts ...score.Option) error {
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

func ParseFile(mainFile path.Local, params []string, opts ...score.Option) (sc *score.Score, err error) {
	fsys, err := dirfs.New(mainFile.Dir())

	if err != nil {
		panic(err.Error())
	}

	opts = append(opts, score.FS(fsys))

	sc = score.New(path.Relative(path.Name(mainFile)), params, opts...)
	err = sc.Parse()
	return
}

func PDFScore(mainFile path.Local, params []string, pdffile path.Local, opts ...score.Option) error {
	fsys, err := dirfs.New(mainFile.Dir())

	if err != nil {
		panic(err.Error())
	}

	opts = append(opts, score.FS(fsys))

	sc := score.New(path.Relative(path.Name(mainFile)), params, opts...)
	err = sc.Parse()
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

	return lilypond.WriteFile(sc, pdffile)
}

func LilypondScore(mainFile path.Local, params []string, opts ...score.Option) (string, error) {
	fsys, err := dirfs.New(mainFile.Dir())

	if err != nil {
		panic(err.Error())
	}

	opts = append(opts, score.FS(fsys))

	sc := score.New(path.Relative(path.Name(mainFile)), params, opts...)
	err = sc.Parse()
	if err != nil {
		return "", err
	}
	err = sc.Format()
	if err != nil {
		return "", err
	}

	err = sc.Unroll()
	if err != nil {
		return "", err
	}

	return lilypond.Create(sc, "*")
}

func Convert(mainFile path.Local, params []string, smffile path.Local, opts ...score.Option) error {
	fsys, err := dirfs.New(mainFile.Dir())

	if err != nil {
		panic(err.Error())
	}

	opts = append(opts, score.FS(fsys))

	sc := score.New(path.Relative(path.Name(mainFile)), params, opts...)
	err = sc.Parse()
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

// func WriteSMFFile(sc *score.Score, smffile string, opts ...smfwriter.Option) error {
func WriteSMFFile(sc *score.Score, smffile path.Local) error {
	fsys, err := rootfs.New()
	if err != nil {
		return err
	}
	return smf.WriteFileFS(sc, fsys, smffile.RootRelative())
}

// WriteImage expects the given score to be unrolled
func WriteImage(sc *score.Score, outfile path.Local, opts ...smfimage.Option) error {
	var midibf bytes.Buffer
	err := smf.WriteSMFTo(sc, &midibf, "*")

	if err != nil {
		return err
	}

	return image.MakeImage(sc, &midibf, outfile.ToSystem(), opts...)
}
