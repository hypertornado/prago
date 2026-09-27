package prago

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func (table *Table) ToExcelFile() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	for k, data := range table.data {
		sheetName := fmt.Sprintf("Sheet%d", k+1)
		if k == 0 {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, err
			}
		} else {
			if _, err := f.NewSheet(sheetName); err != nil {
				return nil, err
			}
		}

		for rowIndex, row := range data.Rows {
			for colIndex, cell := range row.Cells {
				cellName, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
				if err != nil {
					return nil, err
				}
				if err := f.SetCellValue(sheetName, cellName, cell.data.OriginalValue); err != nil {
					return nil, err
				}
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
