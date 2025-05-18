package main

import (
	"os"
	"path/filepath"

	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/muskel"
)

type pdfCmd struct {
	*config.Config
	PDFfile config.FileGetter
}

var PDF = pdfCmd{}

func init() {
	PDF.init()
}

func (p *pdfCmd) init() {
	p.Config = CONFIG.Command("pdf", "convert a muskel file to pdf via lilypond").SkipAllBut("file")
	p.PDFfile = p.File("pdf", "pdf file to export to", config.Required())
}

func (p *pdfCmd) run() error {

	dir, err := os.MkdirTemp(os.TempDir(), "muskel-pdf*")

	if err != nil {
		return err
	}

	file := filepath.Join(dir, "muskel-lilypond.mid")

	defer os.RemoveAll(dir)

	muskel.Convert(ARGS.InFile.Get(), nil, file)

	return lilypond.MIDI2PDF(file, p.PDFfile.Get().ToSystem(), true)
}
