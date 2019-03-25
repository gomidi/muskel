package score

import (
	"os"
	"strings"
)

// pad pads the given string with empty space to the right if it is less then length
func Pad(s string, length int) string {
	if len(s) >= length {
		return s
	}

	diff := length - len(s)
	var bf strings.Builder

	bf.WriteString(s)

	for i := 0; i < diff; i++ {
		bf.WriteString(" ")
	}
	return bf.String()
}

func FileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
