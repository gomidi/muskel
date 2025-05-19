package main

import (
	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/gomidi/lilypond"
	"gitlab.com/gomidi/muskel"
)

type pdfCmd struct {
	*config.Config
	PDFfile      config.FileGetter
	Experimental config.BoolGetter
}

var PDF = pdfCmd{}

func init() {
	PDF.init()
}

func (p *pdfCmd) init() {
	p.Config = CONFIG.Command("pdf", "convert a muskel file to pdf via lilypond").SkipAllBut("file")
	p.PDFfile = p.File("pdf", "pdf file to export to", config.Required())
	p.Experimental = p.Bool("experiment", "use the experimental work in progress lilypond support", config.Shortflag('x'))
}

func (p *pdfCmd) run() error {

	if p.Experimental.Get() {
		return muskel.PDFScore(ARGS.InFile.Get(), nil, p.PDFfile.Get())
	}

	dir, err := rootfs.MkDirTmp("muskel-pdf*")

	if err != nil {
		return err
	}

	file := dir.Join("muskel-lilypond.mid")

	defer rootfs.Delete(dir, true)

	err = muskel.Convert(ARGS.InFile.Get(), nil, file)

	if err != nil {
		return err
	}

	return lilypond.MIDI2PDF(file.ToSystem(), p.PDFfile.Get().ToSystem(), true)
}
