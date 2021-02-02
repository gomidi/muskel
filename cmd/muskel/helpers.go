package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"

	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/midicatdrv"
)

func newDriver() (midi.Driver, error) {
	return driver.New()
}

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
