package score

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"gitlab.com/gomidi/muskel/items"
)

type internalScaleChange struct {
	relPos uint
	scale  *items.Scale
}

type internalScaleChanges []internalScaleChange

func (i internalScaleChanges) Swap(a, b int) {
	i[a], i[b] = i[b], i[a]
}

func (i internalScaleChanges) Less(a, b int) bool {
	return i[a].relPos < i[b].relPos
}

func (i internalScaleChanges) Len() int {
	return len(i)
}

func convertParam(p string) (key string, num int, val string, err error) {
	s := strings.SplitN(p, ":", 2)
	k := strings.SplitN(s[0], "#", 2)
	num, err = strconv.Atoi(k[1])
	if err != nil {
		err = fmt.Errorf("invalid syntax for param %q", p)
		return
	}

	key = k[0]
	val = s[1]
	return
}

func convertParams(in []string) (out map[string][]string, err error) {
	out = map[string][]string{}
	max := map[string]int{}

	for _, p := range in {
		key, num, val, err := convertParam(p)
		if err != nil {
			return nil, err
		}

		if num > max[key] {
			for i := max[key]; i <= num; i++ {
				out[key] = append(out[key], "")
			}
			out[key][num] = val
			max[key] = num
		}
	}

	return
}

type formatter struct {
	wr io.Writer
}

func (f *formatter) WriteLine(line string) error {
	_, err := fmt.Fprintln(f.wr, line)
	return err
}

func (f *formatter) WriteDashLine(line string) error {
	_, err := fmt.Fprintln(f.wr, line)
	return err
}

type sortedLines []line

func (s sortedLines) Len() int {
	return len(s)
}

func (s sortedLines) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

func (s sortedLines) Less(a, b int) bool {
	if s[a].pos < s[b].pos {
		return true
	}

	if s[a].pos > s[b].pos {
		return false
	}

	if len(s[a].cols) > 0 && len(s[a].cols[0]) > 0 && s[a].cols[0][0] == '#' {
		return true
	}

	if len(s[a].cols) == 0 {
		return true
	}

	return false
}

type line struct {
	pos  uint
	cols []string
}
