package table

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/muskel/track"
)

type Tracks struct {
	*Table
}

func NewTracks(lineNo int, sc Score) *Tracks {
	t := &Tracks{
		Table: NewTable("TRACK", lineNo, sc),
	}
	//	sc.tracks = t
	return t
}

func (t *Tracks) Finish() (err error) {
	for i, trackname := range t.Table.Cols() {
		// fmt.Printf("track %q\n", trackname)
		var tr = track.New(trackname)

		for _, line := range t.Table.Data {
			// fmt.Printf("%q: %q\n", line[0], line[i+1])
			if len(line) > i+1 {
				val := line[i+1]
				switch strings.ToLower(line[0]) {
				case "del", "delay":
					err = tr.SetDelay(val)
				case "ch", "channel":
					err = tr.SetChannel(val)
				case "prog", "program":
					err = tr.SetProgram(val)
				case "vol", "volume":
					err = tr.SetVolume(val)
				case "bank":
					err = tr.SetBank(val)
				case "trans", "transpose":
					err = tr.SetTranspose(val)
				case "pbrange", "pitchbendrange":
					err = tr.SetPitchbendrange(val)
				case "velscale", "velocityscale":
					err = tr.SetVelScale(val)
				case "solo":
					err = tr.SetSoloGroup(val)
				case "ambitus":
					err = tr.SetAmbitus(val)
				case "file", "filegroup":
					err = tr.SetFileGroup(val)
				case "import":
					err = tr.SetImport(val)
				default:
					err = fmt.Errorf("unknown property %q in tracks table", line[0])
				}

				if err != nil {
					err = fmt.Errorf("can't set property %q for track %q: %s", line[0], trackname, err.Error())
					return err
				}
			}
		}

		//fmt.Printf("%#v\n", *tr)

		t.Table.Score.AddTrack(tr)
	}
	return nil
}
