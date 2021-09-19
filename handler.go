package xlsx

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	defaultImage = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEUAAACnej3aAAAAAXRSTlMAQObYZgAAAApJREFUCNdjYAAAAAIAAeIhvDMAAAAASUVORK5CYII="
)

func (s *Templater) fieldNameKyeHandler(file *excelize.File, sheet string, rowIdx *int, colIdx int, value interface{}) error {
	axis, _ := excelize.CoordinatesToCellName(colIdx+1, *rowIdx+1)
	file.SetCellValue(sheet, axis, value)
	return nil
}

func (s *Templater) arrayKeyHandler(file *excelize.File, sheet string, rowIdx *int, colIdx int, value interface{}) error {
	rowNumb := *rowIdx + 1
	rows, _ := file.GetRows(sheet)

	array, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("arrayKeyHandler: wrong type payload, array type expected")
	}
	hRowNumb := rowNumb + 1
	hRow := rows[hRowNumb-1]
	for i, item := range array {
		file.DuplicateRowTo(sheet, hRowNumb, hRowNumb+i+1)
		for j := colIdx; j < len(hRow); j++ {
			cellValue := hRow[j]
			placeholderType, value, err := s.placeholder.GetValue(item, cellValue)
			if err != nil {
				return err
			}
			if placeholderType == FieldNameType {
				rowIdx := hRowNumb + i
				s.fieldNameKyeHandler(file, sheet, &rowIdx, j, value)
			}
		}
	}

	file.RemoveRow(sheet, rowNumb)
	file.RemoveRow(sheet, rowNumb)
	if len(array) == 0 {
		file.RemoveRow(sheet, *rowIdx)
		*rowIdx--
	}
	*rowIdx = *rowIdx + len(array) - 2
	return nil
}

func (s *Templater) qrCodeHandler(file *excelize.File, sheet string, rowIdx *int, colIdx int, value interface{}) (err error) {
	colSize, _ := file.GetRowHeight(sheet, *rowIdx+1)
	qrcodeSize := colSize * 1.333
	qrcodeArr, ok := value.([]interface{})
	if !ok {
		err = fmt.Errorf("qrCodeHandler: wrong type payload, array type expected")
		return
	}
	for _, qrcodeStr := range qrcodeArr {
		str, ok := qrcodeStr.(string)
		if !ok {
			err = fmt.Errorf("qrCodeHandler: wrong type payload, string type expected")
			return
		}
		var data []byte
		if data, err = s.qrcodeEncode(str, int(qrcodeSize)); err != nil {
			err = fmt.Errorf("qrCodeHandler: qrcode generate %s", err)
			return
		}
		axis, _ := excelize.CoordinatesToCellName(colIdx+1, *rowIdx+1)
		if err = file.AddPictureFromBytes(sheet, axis, "", "", ".png", data); err != nil {
			err = fmt.Errorf("qrCodeHandler: insert qrcode to file %s", err)
			return
		}
		file.SetCellValue(sheet, axis, "")
		colNum, _, _ := getNumMergeCell(file, sheet, axis)
		colIdx += colNum
	}
	return
}

func (s *Templater) imageHandler(file *excelize.File, sheet string, rowIdx *int, colIdx int, value interface{}) error {
	axis, _ := excelize.CoordinatesToCellName(colIdx+1, *rowIdx+1)

	image, ok := value.(string)
	if !ok {
		return fmt.Errorf("imageHandler: wrong type payload, string type expected")
	}
	i := strings.Index(image, ",")
	image = image[i+1:]
	if len(image) == 0 {
		image = defaultImage
	}
	imageBytes, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return fmt.Errorf("imageHandler: decode image %s", err)
	}
	file.SetCellValue(sheet, axis, "")
	if err := file.AddPictureFromBytes(sheet, axis, "", "", ".png", imageBytes); err != nil {
		return fmt.Errorf("imageHandler: insert image to file %s", err)
	}
	return nil
}

// For quick work add to github.com/xuri/excelize/v2 function:
// GetNumMergeCell provides a function to get the number of merged rows and columns by axis cell
// from a worksheet currently.
// func (f *File) GetNumMergeCell(sheet string, axis string) (colNum int, rowNum int, err error) {
// 	ws, err := f.workSheetReader(sheet)
// 	if err != nil {
// 		return
// 	}

// 	if ws.MergeCells != nil {
// 		for i := range ws.MergeCells.Cells {
// 			ref := ws.MergeCells.Cells[i].Ref
// 			cells := strings.Split(ref, ":")
// 			if cells[0] == axis {
// 				col1, row1, _ := CellNameToCoordinates(cells[0])
// 				col2, row2, _ := CellNameToCoordinates(cells[1])
// 				colNum, rowNum = col2-col1+1, row2-row1+1
// 				return
// 			}
// 		}
// 	}
// 	colNum = 1
// 	rowNum = 1
// 	return
// }
func getNumMergeCell(file *excelize.File, sheet string, axis string) (colNum int, rowNum int, err error) {
	mergedCells, err := file.GetMergeCells(sheet)
	if err != nil {
		return
	}
	for _, mergetCell := range mergedCells {
		if mergetCell.GetStartAxis() == axis {
			col1, row1, _ := excelize.CellNameToCoordinates(mergetCell.GetStartAxis())
			col2, row2, _ := excelize.CellNameToCoordinates(mergetCell.GetEndAxis())
			colNum, rowNum = col2-col1+1, row2-row1+1
			return
		}
	}
	colNum = 1
	rowNum = 1
	return
}