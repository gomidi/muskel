package main

import (
	"bytes"

	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/golang-utils/fs/filesystems/rootfs"
	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/midi/v2/smf"
	"gitlab.com/gomidi/muskel/items"
)

type smfJSONCmd struct {
	*config.Config
	File    config.FileGetter
	OutFile config.FileGetter
}

var SMF_JSON = smfJSONCmd{}

func init() {
	SMF_JSON.init()
}

func (p *smfJSONCmd) init() {
	p.Config = CONFIG.Command("smfjson", "convert a SMF file to a JSON file that can be edited and used as a 'midisample'").SkipAllBut()
	p.OutFile = p.Config.File("outfile", "the json file output file")
	p.File = p.Config.LastFile("file", "the SMF file", config.Required())
}

func (p *smfJSONCmd) run() error {
	bt, err := rootfs.ReadFile(p.File.Get())

	if err != nil {
		return err
	}

	sm, err := smf.ReadFrom(bytes.NewReader(bt))

	if err != nil {
		return err
	}

	bt, err = sm.MarshalJSONIndent()

	if err != nil {
		return err
	}

	var targetFile path.Local

	if p.OutFile.IsSet() {
		targetFile = p.OutFile.Get()
	} else {
		fname := path.Name(p.File.Get())
		targetFile = path.MustWD().Join(items.MIDI_SAMPLES_DIR.String(), fname)
	}

	return rootfs.WriteFile(targetFile, bt, false)
}
