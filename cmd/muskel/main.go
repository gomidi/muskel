package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"gitlab.com/golang-utils/config/v2"
	"gitlab.com/gomidi/muskel"
)

var (
	CONFIG = config.New("muskel",
		muskel.VERSION.Major, muskel.VERSION.Minor, muskel.VERSION.Patch,
		"muskel is a musical sketch language",
		config.AsciiArt("muskel"),
	)
	SIGNAL_CHANNEL = make(chan os.Signal, 10)
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err.Error())
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
		return CLIENT.run(ARGS)
	//case IMPORT.Config:
	//	return IMPORT.run(ARGS)
	case TEMPLATE.Config:
		return TEMPLATE.run()
	default:
		if !ARGS.IgnoreMuskelVersion.Get() {
			// check, if muskel_version.txt is inside the current dir
			// If the version does not match, it delegates to the versioned muskel command (e.g. muskel_1_1_33.exe)
			// checkMuskelVersion exits if appropriate
			checkMuskelVersion(muskel.VERSION.String(), muskel.MUSKEL_VERSION_FILE, ARGS)
		}
	}

	// when using markdown files, set the file extension properly (affects includes!)
	switch filepath.Ext(ARGS.File.Get()) {
	case ".md":
		muskel.FILE_EXTENSION = ".md"
	case ".csv":
		muskel.FILE_EXTENSION = ".csv"
	case ".xlsx":
		muskel.FILE_EXTENSION = ".xlsx"
	default:
		muskel.FILE_EXTENSION = ".mskl"
	}

	// listen for ctrl+c
	go signal.Notify(SIGNAL_CHANNEL, os.Interrupt)

	player, err := newPlayer()
	if err != nil {
		return err
	}

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
