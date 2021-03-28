package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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

func convertArgStringToInts(s string) (res []int) {
	ms := strings.Split(s, ",")

	for _, m := range ms {
		i, errI := strconv.Atoi(strings.TrimSpace(m))
		if errI == nil {
			res = append(res, i)
		}
	}
	return
}
