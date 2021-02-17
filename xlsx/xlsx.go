package xlsx

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
	"gitlab.com/gomidi/muskel/csv"
)

type file struct {
	rd     csv.Reader
	record []string
}

func Read(path string) (mskl string, err error) {

	wb, err := excelize.OpenFile(path)

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
