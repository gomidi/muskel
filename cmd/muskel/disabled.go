package main

// commands that don't support the main argument set

//cmdAddTrack     = cfg.MustCommand("addtrack", "add a track")
//argAddTrackName = cmdAddTrack.NewString("name", "name of the track", config.Required)

//cmdRenameTemplate    = cfg.MustCommand("renametemplate", "rename a template")
//argRenameTemplateOld = cmdRenameTemplate.NewString("old", "old name of the template", config.Required)
//argRenameTemplateNew = cmdRenameTemplate.NewString("new", "new name of the template", config.Required)

//cmdRenameTrack    = cfg.MustCommand("renametrack", "rename a track")
//argRenameTrackOld = cmdRenameTrack.NewString("old", "old name of the track", config.Required)
//argRenameTrackNew = cmdRenameTrack.NewString("new", "new name of the track", config.Required)

//cmdRmTrack     = cfg.MustCommand("rmtrack", "remove a track")
//argRmTrackName = cmdRmTrack.NewString("name", "name of the track that should be removed", config.Required)

//cmdSyncTracks = cfg.MustCommand("synctracks", "sync tracks in include files to the tracks within the given main file")

/*
func runAddTrack() error {
	inFile := argFile.Get()
	sketch := argSketch.Get()

	sc, err := muskel.ParseFile(inFile, sketch)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	name := strings.TrimSpace(argAddTrackName.Get())

	if name == "" {
		return fmt.Errorf("empty name is not allowed")
	}
	trck := score.NewTrack(name)

	sc.AddTrack(trck)

	err = muskel.WriteFormattedFile(sc, inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

/*
func runRmTrack() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	name := strings.TrimSpace(argRmTrackName.Get())
	sc.RmTrack(name)

	err = muskel.WriteFormattedFile(sc, inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

/*
func runSyncTracks() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return muskel.SyncTracks(sc)
}
*/

// RenameTrack(old, nu string)
/*
func runRenameTrack() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	old := strings.TrimSpace(argRenameTrackOld.Get())
	nu := strings.TrimSpace(argRenameTrackNew.Get())

	if nu == "" {
		return fmt.Errorf("empty name is not allowed")
	}

	err = muskel.RenameTrack(sc, old, nu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/

// RenameTrack(old, nu string)
/*
func runRenamePattern() error {
	inFile := argFile.Get()

	sc, err := muskel.ParseFile(inFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	old := strings.TrimSpace(argRenameTemplateOld.Get())
	nu := strings.TrimSpace(argRenameTemplateNew.Get())

	if nu == "" {
		return fmt.Errorf("empty name is not allowed")
	}

	err = muskel.RenameTemplate(sc, old, nu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	return nil
}
*/
