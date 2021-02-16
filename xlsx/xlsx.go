package xlsx

import (
	"fmt"

	"github.com/tealeg/xlsx/v3"
	"gitlab.com/gomidi/muskel/csv"
)

// see tutorial at https://github.com/tealeg/xlsx/blob/master/tutorial/tutorial.adoc

type file struct {
	wb     *xlsx.File
	rd     csv.Reader
	record []string
}

func Read(path string) (mskl string, err error) {
	//fmt.Println("xlsx called")
	// open an existing file
	wb, err := xlsx.OpenFile(path)
	if err != nil {
		return "", err
	}

	f := &file{
		wb: wb,
	}

	// wb now contains a reference to the workbook
	// show all the sheets in the workbook
	//fmt.Println("Sheets in this file:")
	f.eachSheet(func(i int, sh *xlsx.Sheet) {
		//fmt.Println(i, sh.Name)
		//fmt.Println("Max row in sheet:", sh.MaxRow)
		// sh is a reference to a sheet, see above

		/*
			row, err := sh.Row(1)
			if err != nil {
				panic(err)
			}
		*/
		// let's so something with the row ...
		//fmt.Println(row)

		//fmt.Println("Max row is", sh.MaxRow)

		f.rd.WriteString("\n\n## "+sh.Name+"\n\n\n", false)

		err = sh.ForEachRow(f.eachRow)
		if err != nil {
			//return "", err
			fmt.Println("Err=", err)
		}

	})
	s := f.rd.String()
	//fmt.Println(s)
	return s, nil
}

func (f *file) eachSheet(callback func(int, *xlsx.Sheet)) {
	for i, sh := range f.wb.Sheets {
		callback(i, sh)
	}
}

/*
you can use the function GetCoordinate() of the Row struct, which will return an integer with the zero-based row index.



    ColIndexToLetters(index int) – converts a numeric index to the letter combination for a cell address.

    ColLettersToIndex(colLetter string) – converts a column address to a numeric index.

    `GetCoordsFromCellIDString(cellAddr string) – converts a cell address string to row/col coordinates.

    GetCellIDStringFromCoords(x, y int) – converts coordinate values to a cell address

Accessing a single cell can be achieved from the Sheet struct as well as form the Row struct.


Getting a cell from a row

The function GetCell(colIdx int) returns the Cell pointer at a given column index, creating it if it doesn’t exist. That’s the reason there is no error code. If you try to access a cell that is too far "on the right", the package will simply extend the row and create the cell for you.

The Row provides a callback function to iterate over each row in the sheet, ForEachCell().
The argument is a "cell visitor function". This is a function that receives a pointer to a cell as its only parameter and returns an error code. Of course you are free to use an anonymous function, but for the sake of clarity I defined a function named cellVisitor() in the example below. Here’s the complete listing for do a (very simplistic) dump of a sheet from our sample file:
*/

func (f *file) eachRow(r *xlsx.Row) error {
	//fmt.Println(r)
	f.record = nil
	err := r.ForEachCell(f.eachCell)
	if err != nil {
		return err
	}

	//fmt.Printf("record: %#v\n", f.record)
	f.rd.WriteRecord(f.record)

	return nil
}

/*
Starting with v.3.2.0, you can use the function GetCoordinates() of the Cell struct, which will return an integer pair with the zero-based column index and the row index.

Cell types and content
Cell types

The basic data types for an Excel cell are

    Bool

    String

    Formula

    Number

    Date

    Error

    Empty

xlsx.Cell provides a SetXXX() function for the various data types (also splitting numeric data in SetInt(), SetFloat() for instance).

Date values are stored as numeric value with a date format applied. Yes, the list above contains a Date type, but let me quote the comment in the code for this:

Getting cell values

You can retrieve the contents of a cell using these functions

    Value() – returns a string

    FormattedValue() – return a value with the cell’s format applied and an error code

    String() – returns the cell’s value as a string

    Formula() – return a string containing the cell’s formula (or an empty string, if no formula)

    Int() - return the cell’s content as integer and an error code

    Float() - return the cell’s content as a float64 and an error code

    Bool() - return true or false

        if cell has CellTypeBool and the value equals 1, return true

        if cell has CellTypeNumeric and the value is a non-zero, return true

        otherwise return true, if the result of Value() is a non-empty string


*/
func (f *file) eachCell(c *xlsx.Cell) error {
	value := c.Value
	f.record = append(f.record, value)
	/*
		value, err := c.FormattedValue()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Cell value:", value)
		}
	*/
	return nil
}
