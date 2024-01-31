package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/golang-utils/config/v2"
)

type clientCmd struct {
	*config.Config
	ServerAddress config.StringGetter
	Toggle        config.BoolGetter
	Play          config.BoolGetter
	Convert       config.BoolGetter
}

var CLIENT = &clientCmd{}

func init() {
	CLIENT.init()
}

func (s *clientCmd) init() {
	s.Config = CONFIG.Command("client", "client to the server. allows to send play/stop, toggle and convert commands").SkipAllBut("solo", "cutout")
	s.ServerAddress = s.String("addr", "address of the server", config.Default("localhost:8800"))
	s.Toggle = s.Bool("toggle", "toggle between play/stop playing the current file", config.Shortflag('t'))
	s.Play = s.Bool("play", "play (true) or stop (false) playing the current file", config.Shortflag('p'))
	s.Convert = s.Bool("convert", "converts the current file to smf", config.Shortflag('c'))
}

func (c *clientCmd) run(a *args) error {
	var r *http.Response
	var err error
	addr := "http://" + c.ServerAddress.Get() + "/"
	switch {
	case c.Play.IsSet():
		if c.Play.Get() {
			r, err = http.Get(addr + "play")
		} else {
			r, err = http.Get(addr + "stop")
		}
	case c.Convert.IsSet():
		if c.Convert.Get() {
			r, err = http.Get(addr + "convert")
		}
	case c.Toggle.IsSet():
		if c.Toggle.Get() {
			r, err = http.Get(addr + "toggle")
		}
	case a.Solo.IsSet():
		r, err = http.Get(addr + fmt.Sprintf("solo?solo=%v", a.Solo.Get()))
	case a.CutOut.IsSet():
		if a.CutOut.Get() {
			r, err = http.Get(addr + "cutout?cutout=true")
		} else {
			r, err = http.Get(addr + "cutout?cutout=false")
		}
	default:
		return fmt.Errorf("pass either --play, --toogle or --convert or --solo or --cutout")
	}

	if r != nil {
		ioutil.ReadAll(r.Body)
		r.Body.Close()
	}

	return err
}
