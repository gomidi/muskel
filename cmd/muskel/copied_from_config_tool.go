package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

/*
the following code is copied (due to missing internet connection) from
gitlab.com/metakeule/config
and should eventuelly be used from there directly
*/

func Versionate(file string, v *version) (file_versionated string) {
	return fmt.Sprintf("%s_%v_%v_%v", file, v.major, v.minor, v.patch)
}

const VERSION_FILE = "VERSION"

func readVERSIONFile(wd string) (v *version, err error) {
	res, err1 := ioutil.ReadFile(filepath.Join(wd, VERSION_FILE))
	if err1 != nil {
		return nil, err1
	}

	v, err = parseVersion(strings.TrimSpace(string(res)))
	//fmt.Printf("parsing from VERSION file: %v\n", v)
	if err != nil {
		v = nil
	}

	return
}

type version struct {
	major int
	minor int
	patch int
}

func (v version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.major, v.minor, v.patch)
}

func parseVersion(v string) (*version, error) {
	v = strings.TrimLeft(v, "v")
	nums := strings.Split(v, ".")
	var ns = make([]int, len(nums))

	var ver version

	for i, n := range nums {
		nn, err := strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
		ns[i] = nn
	}

	switch len(ns) {
	case 1:
		ver.major = ns[0]
	case 2:
		ver.major = ns[0]
		ver.minor = ns[1]
	case 3:
		ver.major = ns[0]
		ver.minor = ns[1]
		ver.patch = ns[2]
	default:
		return nil, fmt.Errorf("invalid version string %q", v)
	}

	return &ver, nil

}

type versions []*version

func (v versions) Less(a, b int) bool {
	if v[a].major != v[b].major {
		return v[a].major < v[b].major
	}

	if v[a].minor != v[b].minor {
		return v[a].minor < v[b].minor
	}

	return v[a].patch < v[b].patch
}

func (v versions) Len() int {
	return len(v)
}

func (v versions) Swap(a, b int) {
	v[a], v[b] = v[b], v[a]
}

func (v versions) Last() *version {
	if len(v) == 0 {
		return nil
	}

	return v[len(v)-1]
}
