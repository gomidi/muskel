package main

import (
	"fmt"
	"os"

	"gitlab.com/golang-utils/fs"
	"gitlab.com/golang-utils/fs/filesystems/mapfs"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	lpond "gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/muskel/lilypond"
	"gitlab.com/gomidi/muskel/score"
)

func main() {
	err := run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}

}

func run() error {

	fsys, err := mapfs.New(path.MustLocal("/hiho/"))

	if err != nil {
		return err
	}

	file := path.Relative("main.md")

	err = writeMainFile(fsys, file)

	if err != nil {
		return err
	}

	lbk, err := getLilypondScore(fsys, file)

	if err != nil {
		return err
	}

	err = writeLilyPondScore(lbk.String(), path.MustWD().Join("main.ly"))

	if err != nil {
		return err
	}

	//return nil
	err = writeLilyPondPDF(lbk, path.MustWD().Join("main.pdf"))
	return err
}

func writeLilyPondPDF(bk *lpond.Book, file path.Local) error {
	return bk.ToPDF(file)
}

func writeLilyPondScore(s string, file path.Local) error {
	return rootfs.WriteFile(file, []byte(s), false)
}

func writeMainFile(fsys fs.FS, file path.Relative) error {
	return fs.WriteFile(fsys, file, []byte(ScoreData), false)
}

func getLilypondScore(fsys fs.FS, file path.Relative) (*lpond.Book, error) {
	var opts []score.Option

	opts = append(opts, score.FS(fsys))

	sc := score.New(file, nil, opts...)
	err := sc.Parse()
	if err != nil {
		fmt.Printf("parsing error: %s\n", err.Error())
		return nil, err
	}

	/*
		err = sc.Format()
		if err != nil {
			return nil, err
		}
	*/

	err = sc.Unroll()
	if err != nil {
		return nil, err
	}

	return lilypond.CreateBook(sc, "*")
}

var ScoreData = `

| TRACK      | Vocal | Violin | Bass |
| ---------- | ----- | ------ | ---- |
|  Channel   | 1     | 2      | 3    |
|  Transpose | 12    | 0      | 0    |

| @lyrics       |       |
| ------------- | ----- |
|  _  who      | verse |
|  se-pa-ra-ted |       |

| =motiv   |     | violin  |
| -------- | --- | ------- |
|  # 3/4   |     |         |
|     1    | *   | c       |
|     1&   |     |         |
|     2    | d   | (c e g) |
|     2&   | e   | d       |
|     3    | *   | (d f a) |
|     3&   | g   | c       |
|  #       |     |         |
|     1    |     |         |
|     1&   | a:: | c       |
|     2&   | d   | d:      |
|     3    | e:: | c       |
|     3&   | f   | d       |


| =SCORE   | Vocal                  | Violin           | Bass      |
| -------- | ---------------------- | ---------------- | --------- |
|  # 3/4   |                        |                  |           |
|     1    | =motiv%12@lyrics.verse |                  |           |
|     1&   |                        |                  |           |
|     2    |                        |                  |           |
|     2&   |                        | =motiv.violin%12 |           |
|     3    |                        |                  |           |
|     3&   |                        |                  |           |
|  #       |                        |                  |           |
|     1    |                        |                  | =motiv%12 |
|     1&   |                        |                  |           |
|     2    |                        |                  |           |
|     2&   |                        |                  |           |
|     3    |                        |                  |           |
|     3&   |                        |                  |           |
|  #       |                        |                  |           |
|     1    |                        |                  |           |
|     1&   |                        |                  |           |
|     2    |                        |                  |           |
|     2&   |                        |                  |           |
|     3    |                        |                  |           |
|     3&   |                        |                  |           |
|  #       |                        |                  |           |
|  *5      |                        |                  |           |
|  #       |                        |                  |           |
|     1    | *                      | *                | *         |


`
