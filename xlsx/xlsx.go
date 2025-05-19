package xlsx

import (
	"fmt"
	"io"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"gitlab.com/gomidi/muskel/csv"
)

type file struct {
	rd     csv.Reader
	record []string
}

func Write(path string, tracksTable string, scoreTable string) error {
	xl := excelize.NewFile()
	xl.SetSheetName("Sheet1", "TRACK")

	trLines := strings.Split(tracksTable, "\n")
	writeLines(xl, "TRACK", trLines)

	sc := xl.NewSheet("SCORE")
	scLines := strings.Split(scoreTable, "\n")
	writeLines(xl, "SCORE", scLines)

	xl.SetActiveSheet(sc)

	return xl.SaveAs(path)
}

func writeLines(xl *excelize.File, sheet string, lines []string) {
	rcorr := 0

	for r, line := range lines {
		if isMarkdownDashLine(line) {
			rcorr--
			continue
		}

		cols := strings.Split(line, "|")
		if len(cols) > 1 {
			for c, col := range cols[1:] {
				xl.SetCellStr(sheet, positionToAxis(r+rcorr, c), strings.TrimSpace(col))
			}
		} else {
			xl.SetCellStr(sheet, positionToAxis(r+rcorr, 0), strings.TrimSpace(line))
		}
	}

}

func ReadFrom(rd io.Reader) (mskl string, err error) {
	wb, err := excelize.OpenReader(rd)

	if err != nil {
		return "", err
	}

	f := &file{}

	no := wb.SheetCount

	for i := 0; i < no; i++ {
		name := wb.GetSheetName(i + 1)
		rows := wb.GetRows(name)
		f.rd.WriteString("\n\n## "+name+"\n\n\n", false)

		if err != nil {
			//return "", err
			fmt.Println("Err=", err)
		}

		for _, row := range rows {
			f.record = nil

			//fmt.Printf("record: %#v\n", f.record)

			for _, colCell := range row {
				value := colCell
				f.record = append(f.record, value)
				//fmt.Print(colCell, "\t")
			}
			//fmt.Println()
			f.rd.WriteRecord(f.record)
		}
	}

	s := f.rd.String()
	//fmt.Println(s)
	return s, nil
}
