package xlsx

import (
	"fmt"
	"strconv"
	"strings"
)

func isMarkdownDashLine(line string) bool {
	if !strings.Contains(line, "--") {
		return false
	}

	res := strings.ReplaceAll(line, "|", " ")
	res = strings.ReplaceAll(res, "-", " ")

	return strings.TrimSpace(res) == ""
}

// the following functions are stolen from https://github.com/360EntSecGroup-Skylar/excelize/issues/132

func positionToAxis(row, col int) string {
	if row < 0 || col < 0 {
		return ""
	}
	rowString := strconv.Itoa(row + 1)
	colString := ""
	col++
	for col > 0 {
		colString = fmt.Sprintf("A%v",col%26-1) + colString
		col /= 26
	}
	return colString + rowString
}

func axisToPosition(axis string) (int, int, error) {
	col := 0
	for i, char := range axis {
		if char >= 'A' && char <= 'Z' {
			col *= 26
			col += int(char - 'A' + 1)
		} else if char >= 'a' && char <= 'z' {
			col *= 26
			col += int(char - 'a' + 1)
		} else {
			row, err := strconv.Atoi(axis[i:])
			return row - 1, col - 1, err
		}
	}
	return -1, -1, fmt.Errorf("invalid axis format %s", axis)
}
