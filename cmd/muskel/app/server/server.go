package server

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"gitlab.com/golang-utils/fs/path"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/config"
	"gitlab.com/gomidi/muskel/cmd/muskel/app/oshelpers"
	"gitlab.com/gomidi/muskel/score"
)

type Converter interface {
	ParseMuskel(srcFile path.Local) (*score.Score, error)
	CmdSMF(*score.Score) error
}

type Player interface {
	Play()
	Stop()
	Toggle()
}

type Server struct {
	converter Converter
	player    Player
	inFile    path.Local
	config    config.Config
	mx        sync.Mutex
	file      string
}

func New(conv Converter, conf config.Config, player Player, inFile path.Local, file string) *Server {
	return &Server{
		converter: conv,
		config:    conf,
		player:    player,
		inFile:    inFile,
		file:      file,
	}
}

func (me *Server) Start(addr string) error {
	if !me.config.XLSX && me.config.CSV == "" {
		me.config.Fmt = true
	}
	if me.config.UnrollFile == "" {
		me.config.UnrollFile = me.inFile.String() + ".unrolled" + muskel.FILE_EXTENSION
	}
	me.config.Flow = true
	me.config.KeepEmptyLines = true
	me.config.Debug = false

	http.HandleFunc("/play", me.Play)
	http.HandleFunc("/stop", me.Stop)
	http.HandleFunc("/toggle", me.Toggle)
	http.HandleFunc("/convert", me.Convert)
	http.HandleFunc("/solo", me.SoloGroup)
	http.HandleFunc("/cutout", me.CutOut)
	fmt.Printf("starting server on %q\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (me *Server) setSoloGroup(vals url.Values) error {
	me.mx.Lock()
	defer me.mx.Unlock()
	gr := vals.Get("solo")
	if len(gr) == 0 {
		me.config.SoloGroup = 0
		return nil
	}
	i, err := strconv.Atoi(gr)
	if err != nil {
		return err
	}
	if i <= 0 {
		me.config.SoloGroup = 0
		return nil
	}
	me.config.SoloGroup = uint(i)
	return nil
}

func (me *Server) SoloGroup(wr http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	err := me.setSoloGroup(vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	wr.Write([]byte("solo set"))
}

func (me *Server) setCutout(vals url.Values) error {
	me.mx.Lock()
	defer me.mx.Unlock()
	gr := vals.Get("cutout")
	if len(gr) == 0 {
		me.config.SoloGroup = 0
		return nil
	}
	b, err := strconv.ParseBool(gr)
	if err != nil {
		return err
	}
	me.config.CutOut = b
	return nil
}

func (me *Server) CutOut(wr http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	err := me.setCutout(vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	wr.Write([]byte("cutout set"))
}

func (me *Server) Play(wr http.ResponseWriter, r *http.Request) {
	err := me.convert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	me.player.Play()
	wr.Write([]byte("start playing"))
}

func (me *Server) Stop(wr http.ResponseWriter, r *http.Request) {
	me.player.Stop()
	wr.Write([]byte("stopped playing"))
}

func (me *Server) Toggle(wr http.ResponseWriter, r *http.Request) {
	err := me.convert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	me.player.Toggle()
	wr.Write([]byte("toggle playing"))
}

func (me *Server) convert() error {
	me.mx.Lock()
	defer me.mx.Unlock()
	sc, err := me.converter.ParseMuskel(me.inFile)
	if err != nil {
		oshelpers.Alert("ERROR: ", err)
		return err
	}
	err = me.converter.CmdSMF(sc)
	if err != nil {
		oshelpers.Alert("ERROR: ", err)
		return err
	}
	return nil
}

func (me *Server) Convert(wr http.ResponseWriter, r *http.Request) {
	err := me.convert()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("error: " + err.Error()))
		return
	}
	wr.Write([]byte("converted " + me.file))
}
