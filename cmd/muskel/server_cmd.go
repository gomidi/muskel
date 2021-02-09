package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"gitlab.com/metakeule/config"
)

type server struct {
	mx        sync.Mutex
	converter *converter
}

func (s *server) setSoloGroup(vals url.Values) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	gr := vals.Get("solo")
	if len(gr) == 0 {
		s.converter.Config.SoloGroup = 0
		return nil
	}
	i, err := strconv.Atoi(gr)
	if err != nil {
		return err
	}
	if i <= 0 {
		s.converter.Config.SoloGroup = 0
		return nil
	}
	s.converter.Config.SoloGroup = uint(i)
	return nil
}

func (s *server) SoloGroup(wr http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	err := s.setSoloGroup(vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	wr.Write([]byte("solo set"))
}

func (s *server) setCutout(vals url.Values) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	gr := vals.Get("cutout")
	if len(gr) == 0 {
		s.converter.Config.SoloGroup = 0
		return nil
	}
	b, err := strconv.ParseBool(gr)
	if err != nil {
		return err
	}
	s.converter.Config.CutOut = b
	return nil
}

func (s *server) CutOut(wr http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	err := s.setCutout(vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	wr.Write([]byte("cutout set"))
}

func (s *server) Play(wr http.ResponseWriter, r *http.Request) {
	err := s.convert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	s.converter.player.playCh <- true
	wr.Write([]byte("start playing"))
}

func (s *server) Stop(wr http.ResponseWriter, r *http.Request) {
	s.converter.player.playCh <- false
	<-s.converter.player.stoppedCh
	wr.Write([]byte("stopped playing"))
}

func (s *server) Toggle(wr http.ResponseWriter, r *http.Request) {
	err := s.convert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	s.converter.player.toggleCh <- true
	wr.Write([]byte("toggle playing"))
}

func (s *server) convert() error {
	s.mx.Lock()
	defer s.mx.Unlock()
	sc, err := s.converter.parseMuskel(s.converter.inFile)
	if err != nil {
		alert("ERROR: ", err)
		return err
	}
	err = s.converter.cmdSMF(sc)
	if err != nil {
		alert("ERROR: ", err)
		return err
	}
	return nil
}

func (s *server) Convert(wr http.ResponseWriter, r *http.Request) {
	err := s.convert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	wr.Write([]byte("converted " + s.converter.file))
}

func startServer(conv *converter) error {
	s := &server{
		converter: conv,
	}

	s.converter.setFromArgs(ARGS)
	s.converter.Config.Fmt = true
	s.converter.Config.UnrollFile = s.converter.inFile + ".unrolled.mskl"
	s.converter.Config.Flow = true
	s.converter.Config.KeepEmptyLines = true
	s.converter.Config.Debug = false

	addr := SERVER.Address.Get()

	http.HandleFunc("/play", s.Play)
	http.HandleFunc("/stop", s.Stop)
	http.HandleFunc("/toggle", s.Toggle)
	http.HandleFunc("/convert", s.Convert)
	http.HandleFunc("/solo", s.SoloGroup)
	http.HandleFunc("/cutout", s.CutOut)
	fmt.Printf("starting server on %q\n", addr)
	return http.ListenAndServe(addr, nil)
}

type serverCmd struct {
	*config.Config
	Address      config.StringGetter
	PlayProgram  config.StringGetter
	playershared *Player
}

var SERVER = serverCmd{}

func init() {
	SERVER.init()
}

func (s *serverCmd) init() {
	s.Config = CONFIG.MustCommand("server", "start command server").Skip("watch").Skip("dir")
	s.Address = s.NewString("addr", "address of the server", config.Default("localhost:8800"))
	s.PlayProgram = s.NewString("cmd", "command to execute when playing (fluidsynth,timidity,audacious,auto or midi-ports (port:4) or custom (pass $_file variable))", config.Default("auto"))
}

func (s *serverCmd) run(cbr *converter) error {
	if cbr.player == nil {
		return fmt.Errorf("player must not be nil")
	}

	var stopFunc = func() {
		cbr.player.stopPlayer <- true
		<-cbr.player.playerStopped
	}

	if cbr.player.playToPort {
		go cbr.player.playThroughPort()
	} else {
		go cbr.player.playWithProgram()
	}

	go func() {
		err := startServer(cbr)
		if err != nil {
			fmt.Printf("ERROR: can't start server: %s\n", err.Error())
			stopFunc()
			os.Exit(1)
		}
	}()

	// ctrl+c
	<-SIGNAL_CHANNEL
	fmt.Println("\n--interrupted!")
	stopFunc()
	os.Exit(0)

	return nil
}
