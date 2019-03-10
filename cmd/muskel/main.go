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
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/metakeule/observe/lib/runfunc"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/muskel"
	"gitlab.com/metakeule/config"
)

var (
	cfg = config.MustNew("muskel", "0.9.0", "muskel is a musical sketch language")

	argFile      = cfg.NewString("file", "path of the muskel file", config.Shortflag('f'), config.Required)
	argFmt       = cfg.NewBool("fmt", "format the muskel file (overwrites the input file)")
	argWatch     = cfg.NewBool("watch", "watch for changes of the file and act on each change", config.Shortflag('w'))
	argDir       = cfg.NewBool("dir", "watch for changes in the current directory (not just for the input file)", config.Shortflag('d'))
	argOutFile   = cfg.NewString("out", "path of the output file (SMF)", config.Shortflag('o'))
	argSleep     = cfg.NewInt32("sleep", "sleeping time between invocations (in milliseconds)", config.Default(int32(100)))
	argSmallCols = cfg.NewBool("small", "small columns in formatting", config.Shortflag('s'), config.Default(false))
	argUnroll    = cfg.NewString("unroll", "unroll the source to the given file name", config.Shortflag('u'))
	argDebug     = cfg.NewBool("debug", "print debug messages")

	cmdSMF      = cfg.MustCommand("smf", "convert a muskel file to Standard MIDI file format (SMF)")
	argSMFTicks = cmdSMF.NewInt32("ticks", "resolution of SMF file in ticks", config.Default(int32(960)))

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

func fmtFile(file string) error {
	sc, err := muskel.ParseFile(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
		beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		beeep.Alert("ERROR while parsing MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}

	if argSmallCols.Get() {
		sc.SmallColumns = true
	}

	sc.AddMissingProperties()

	err = sc.WriteToFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
		beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
		return err
	}
	fmt.Printf(".")
	return nil
	//beeep.Notify("OK MuSkeL formatted:", "", "assets/information.png")
}

var lock sync.RWMutex

func runCmd() (callback func(dir, file string) error, file_, dir_ string) {

	if argDebug.Get() {
		muskel.DEBUG = true
	}

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

	// oldFileChecksum := ""
	inFilep, _ := filepath.Abs(inFile)

	dir_ = path.Dir(inFile)
	file_ = path.Base(inFile)

	checksums := map[string]string{}
	ignore := map[string]bool{}

	callback = func(dir, file string) error {
		if filepath.Ext(file) != ".mskl" {
			return nil
		}

		if strings.Contains(dir, "muskel-fmt") || strings.Contains(file, "muskel-fmt") {
			return nil
		}

		filep, _ := filepath.Abs(file)
		if strings.Contains(filep, "muskel-fmt") {
			return nil
		}

		//lock.RLock()
		if ignore[filep] {
			//lock.RUnlock()
			return nil
		}

		// fmt.Printf("[%q, %q]\n", dir, file)

		if !argDir.Get() {
			if filep != inFilep {
				return nil
			}
		}

		//	lock.Lock()
		newChecksum := fileCheckSum(filep)
		//lock.Unlock()

		if newChecksum == checksums[filep] {
			return nil
		}

		//		lock.Lock()
		checksums[filep] = newChecksum
		//	lock.Unlock()

		if argDir.Get() {
			if filep != inFilep {

				//				go func() {
				//fmt.Printf("formatting %q...\n", filep)
				//lock.Lock()
				ignore[filep] = true
				//lock.Unlock()

				fmtFile(filep)
				//lock.Lock()
				ignore[filep] = false
				//lock.Unlock()
				//fmt.Printf("done formatting %q...\n", filep)
				//			}()

			}
		}

		//lock.Lock()
		//ignore[filep] = true
		defer func() {
			//ignore[filep] = false
			//lock.Unlock()
		}()
		sc, err := muskel.ParseFile(inFile)

		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while parsing MuSkeL: %s\n", err.Error())
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
				fmt.Fprintf(os.Stderr, "ERROR while unrolling MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while unrolling MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}

			err = ur.WriteToFile(argUnroll.Get())
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR while writing unrolling MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while writing unrolled MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}

		}

		if argFmt.Get() {
			err = sc.WriteToFile(inFilep)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR while writing formatting MuSkeL: %s\n", err.Error())
				beeep.Alert("ERROR while writing formatting MuSkeL:", err.Error(), "assets/warning.png")
				return err
			}

			//beeep.Notify("OK MuSkeL formatted:", "", "assets/information.png")
		}

		ext := path.Ext(outFile)
		outFile = strings.Replace(outFile, ext, ".mid", -1)

		switch cfg.ActiveCommand() {
		case cmdSMF:
			err = sc.WriteSMF(outFile, smfwriter.TimeFormat(smf.MetricTicks(argSMFTicks.Get())))
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
				beeep.Alert("ERROR while converting MuSkeL to SMF", err.Error(), "assets/warning.png")
				return err
			}

			fmt.Fprint(os.Stdout, ".")
			beeep.Notify("OK MuSkeL converted to SMF", path.Base(outFile), "assets/information.png")
			return nil
		case cmdPlay:
			err = sc.WriteSMF(outFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR while converting MuSkeL to SMF: %s\n", err.Error())
				beeep.Alert("ERROR while converting MuSkeL to SMF", err.Error(), "assets/warning.png")
				return err
			}

			fmt.Fprint(os.Stdout, ".")
			beeep.Notify("OK MuSkeL converted to SMF", path.Base(outFile), "assets/information.png")
			if audacious != "" {
				go func() {
					cmd := exec.Command("audacious", "-1", "-H", "-p", "-q", outFile)
					_, _ = cmd.CombinedOutput()
				}()
				return nil
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
		if argDir.Get() {
			fmt.Printf("watching %q\n", dir)
		} else {
			fmt.Printf("watching %q\n", argFile.Get())
		}
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
