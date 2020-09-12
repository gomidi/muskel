package muskel

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

func Versionate(file string, v *Version) (file_versionated string) {
	return fmt.Sprintf("%s_%v_%v_%v", file, v.Major, v.Minor, v.Patch)
}

const VERSION_FILE = "VERSION"

func ReadVERSIONFile(wd string) (v *Version, err error) {
	res, err1 := ioutil.ReadFile(filepath.Join(wd, VERSION_FILE))
	if err1 != nil {
		return nil, err1
	}

	v, err = ParseVersion(strings.TrimSpace(string(res)))
	//fmt.Printf("parsing from VERSION file: %v\n", v)
	if err != nil {
		v = nil
	}

	return
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}

func ParseVersion(v string) (*Version, error) {
	v = strings.TrimLeft(v, "v")
	nums := strings.Split(v, ".")
	var ns = make([]int, len(nums))

	var ver Version

	for i, n := range nums {
		nn, err := strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
		ns[i] = nn
	}

	switch len(ns) {
	case 1:
		ver.Major = ns[0]
	case 2:
		ver.Major = ns[0]
		ver.Minor = ns[1]
	case 3:
		ver.Major = ns[0]
		ver.Minor = ns[1]
		ver.Patch = ns[2]
	default:
		return nil, fmt.Errorf("invalid version string %q", v)
	}

	return &ver, nil

}

type Versions []*Version

func (v Versions) Less(a, b int) bool {
	if v[a].Major != v[b].Major {
		return v[a].Major < v[b].Major
	}

	if v[a].Minor != v[b].Minor {
		return v[a].Minor < v[b].Minor
	}

	return v[a].Patch < v[b].Patch
}

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Swap(a, b int) {
	v[a], v[b] = v[b], v[a]
}

func (v Versions) Last() *Version {
	if len(v) == 0 {
		return nil
	}

	return v[len(v)-1]
}
