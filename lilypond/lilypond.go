package lilypond

import "gitlab.com/gomidi/lilypond"

func makeLily() *lilypond.Book {
	bk := lilypond.NewBook(lilypond.Staff{})

	// bk.Header.Add()
	layout := bk.Score.Layout.Add()
	_ = layout
	return bk
}
