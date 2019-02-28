package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/metakeule/config"
)

var (
	cfg = config.MustNew("muskel", "0.0.3", "muskel is a musical sketch language")

	argFile      = cfg.NewString("file", "path of the muskel file", config.Shortflag('f'), config.Required)
	argFmt       = cfg.NewBool("fmt", "format the muskel file (overwrites the input file)")
	argWatch     = cfg.NewBool("watch", "watch for changes of the file and act on each change", config.Shortflag('w'))
	argOutFile   = cfg.NewString("out", "path of the output file (SMF)", config.Shortflag('o'))
	argSleep     = cfg.NewInt32("sleep", "sleeping time between invocations (in milliseconds)", config.Default(int32(200)))
	argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
	argUnroll    = cfg.NewString("unroll", "unroll the source to the given file name", config.Shortflag('u'))
	argDebug     = cfg.NewBool("debug", "print debug messages", config.Shortflag('d'))

	cmdSMF  = cfg.MustCommand("smf", "convert a muskel file to Standard MIDI file format (SMF)")
	cmdPlay = cfg.MustCommand("play", "play a muskel file (currently linux only, needs audacious)")
)

func fileCheckSum(file string) string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func runCmd() (callback func(dir, file string) error, file_, dir_ string) {

	inFile := argFile.Get()
	outFile := inFile
	if argOutFile.IsSet() {
		outFile = argOutFile.Get()
	}

	cmd := exec.Command("which", "audacious")
	out, errC := cmd.CombinedOutput()

	audacious := strings.TrimSpace(string(out))

	if errC != nil || audacious == "" {
		audacious = ""
	}

	oldFileChecksum := ""
	inFilep, _ := filepath.Abs(inFile)

	dir_ = path.Dir(inFile)
	file_ = path.Base(inFile)

	callback = func(dir, file string) error {
		filep, _ := filepath.Abs(file)
		if filep != inFilep {
			return nil
		}

		newChecksum := fileCheckSum(filep)

		if newChecksum == oldFileChecksum {
			return nil
		}
		oldFileChecksum = newChecksum

		sc, err := muskel.ParseFile(inFile)

		if err != nil {
			fmt.Printf("ERROR while parsing MuSkeL: %s\n", err.Error())
			beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
			beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
			return err
		}

		if argSmallCols.Get() {
			sc.SmallColumns = true
		}

		sc.AddMissingProperties()

		if argUnroll.IsSet() {
			var ur *muskel.Score
			ur, err = sc.Unroll()
			if err != nil {
				fmt.Printf("ERROR while unrolling MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while unrolling MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}

			err = ur.WriteToFile(argUnroll.Get())
			if err != nil {
				fmt.Printf("ERROR while writing unrolling MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while writing unrolled MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}

		}

		if argFmt.Get() {
			err = sc.WriteToFile(inFilep)
			if err != nil {
				fmt.Printf("ERROR while writing formatting MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}

			//beeep.Notify("OK MuSkeL formatted:", "", "assets/information.png")
		}

		ext := path.Ext(outFile)
		outFile = strings.Replace(outFile, ext, ".mid", -1)

		switch cfg.ActiveCommand() {
		case cmdSMF:
			err = sc.WriteSMF(outFile, argDebug.Get())
			if err != nil {
				fmt.Printf("ERROR while converting MuSkeL to SMF: %s\n", err.Error())
				beeep.Alert("ERROR while converting MuSkeL to SMF", err.Error(), "assets/warning.png")
				return err
			}

			beeep.Notify("OK MuSkeL converted to SMF", path.Base(outFile), "assets/information.png")
			return nil
		case cmdPlay:
			err = sc.WriteSMF(outFile, argDebug.Get())
			if err != nil {
				fmt.Printf("ERROR while converting MuSkeL to SMF: %s\n", err.Error())
				beeep.Alert("ERROR while converting MuSkeL to SMF", err.Error(), "assets/warning.png")
				return err
			}

			beeep.Notify("OK MuSkeL converted to SMF", path.Base(outFile), "assets/information.png")
			if audacious != "" {
				cmd := exec.Command("audacious", "-1", "-H", "-p", "-q", outFile)
				_, err = cmd.CombinedOutput()
				return err
			}
			//beeep.Alert("audacious not found", err.Error(), "assets/warning.png")
			return fmt.Errorf("audacious not found")
		default:
			return nil
		}
	}

	return
	//fmt.Printf("watching: %#v, file changed: %#v\n", dir, file)
	//return fmt.Errorf("test-error\n")
}

func run() error {

	err := cfg.Run()

	if err != nil {
		fmt.Println(cfg.Usage())
		return err
	}

	cmd, file, dir := runCmd()

	if argWatch.Get() {
		fmt.Printf("watching \n%s\n", argFile.Get())
		r := runfunc.New(dir, cmd, runfunc.Sleep(time.Millisecond*time.Duration(int(argSleep.Get()))))

		errors := make(chan error, 1)
		stopped, err := r.Run(errors)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		go func() {
			for e := range errors {
				fmt.Fprintf(os.Stderr, "error: %s\n", e)
				_ = e
			}
		}()

		sigchan := make(chan os.Signal, 10)

		// listen for ctrl+c
		go signal.Notify(sigchan, os.Interrupt)

		// interrupt has happend
		<-sigchan
		fmt.Println("\n--interrupted!")

		stopped.Kill()

		os.Exit(0)
		return nil
	} else {
		return cmd(dir, file)
	}

}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}
