package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"gitlab.com/gomidi/muskel"
	"gitlab.com/metakeule/config"
)

var (
	CONFIG         = config.MustNew("muskel", muskel.VERSION, "muskel is a musical sketch language")
	SIGNAL_CHANNEL = make(chan os.Signal, 10)
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {

	err := CONFIG.Run()

	if err != nil {
		fmt.Println(CONFIG.Usage())
		return err
	}

	// handle commands that do not make use of neither the converter nor the player
	switch CONFIG.ActiveCommand() {
	case PORTS.Config:
		return PORTS.run()
	case CONFIG_DIRS.Config:
		return CONFIG_DIRS.run()
	case CLIENT.Config:
		return CLIENT.run()
	case IMPORT.Config:
		return IMPORT.run()
	case TEMPLATE.Config:
		return TEMPLATE.run()
	default:
		// when using markdown files, set the file extension properly (affects includes!)
		if filepath.Ext(ARGS.File.Get()) == ".md" {
			muskel.FILE_EXTENSION = ".md"
		}
	}

	var player *Player

	if CONFIG.ActiveCommand() == PLAY.Config || CONFIG.ActiveCommand() == SERVER.Config {
		player, err = newPlayer()
		if err != nil {
			return err
		}
	}

	if !ARGS.IgnoreMuskelVersion.Get() {
		// check, if muskel_version.txt is inside the current dir
		// If the version does not match, it delegates to the versioned muskel command (e.g. muskel_1_1_33.exe)
		// checkMuskelVersion exits if appropriate
		checkMuskelVersion(muskel.VERSION, muskel.MUSKEL_VERSION_FILE, ARGS)
	}

	// listen for ctrl+c
	go signal.Notify(SIGNAL_CHANNEL, os.Interrupt)

	conv := newConverter(player, ARGS)

	if CONFIG.ActiveCommand() == SERVER.Config {
		// the server keeps running, until ctrl+c
		return SERVER.run(conv)
	}

	if ARGS.Watch.IsSet() && ARGS.Watch.Get() {
		// the watcher keeps running, until ctrl+c
		return watch(conv, ARGS)
	}

	// just one simple conversion / play; clean exit on ctrl+c
	return conv.run()
}
