package table

import (
	"strings"
	"unicode/utf8"
)

func LenUTF8(s string) int {
	return utf8.RuneCountInString(s)
}

// pad pads the given string with empty space to the right if it is less then length
func Pad(s string, length int) string {
	l := LenUTF8(s)
	if l >= length {
		return s
	}

	diff := length - l
	var bf strings.Builder

	bf.WriteString(s)

	for i := 0; i < diff; i++ {
		bf.WriteString(" ")
	}
	return bf.String()
}

func Dash(l int, length int) string {
	s := strings.Repeat("-", l)
	if l >= length {
		return s
	}

	diff := length - l
	var bf strings.Builder

	bf.WriteString(s)

	for i := 0; i < diff; i++ {
		bf.WriteString(" ")
	}
	return bf.String()
}

func tableHeader(line string) (cols []string) {
	c := strings.Split(line, "|")
	//name = strings.TrimRight(strings.TrimSpace(strings.ToLower(c[0])), ":")
	for i := 2; i < len(c)-1; i++ {
		cols = append(cols, strings.TrimSpace(c[i]))
	}
	//fmt.Printf("cols: %q\n", cols)

	/*
		if len(c) > 0 && strings.TrimSpace(c[len(c)-1]) != "" {
			cols = append(cols, strings.TrimSpace(c[len(c)-1]))
		}
	*/

	return
}

type formatLine string

func (f *formatLine) WriteLine(l string) error {
	*f = formatLine(l)
	return nil
}

func (f *formatLine) WriteDashLine(l string) error {
	*f = formatLine(l)
	return nil
}

var _f formatLine
var _ Formatter = &_f
