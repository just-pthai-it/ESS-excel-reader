package excel_file_reader

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"regexp"
	"strings"
)

type aExcelFileReader struct {
	filePath string
}

func (excelFileReader aExcelFileReader) ReadFirstSheet() ([][]string, error) {
	f, err := excelize.OpenFile(excelFileReader.filePath)
	sheetsList := f.GetSheetList()

	if err != nil {
		fmt.Println(err)
		return [][]string{}, err
	}

	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err2 := f.GetRows(sheetsList[0])
	if err2 != nil {
		fmt.Println(err2)
		return [][]string{}, err2
	}

	return rows, nil
}

func (excelFileReader aExcelFileReader) ReadAllSheets() ([][][]string, error) {
	f, err := excelize.OpenFile(excelFileReader.filePath)
	sheetsList := f.GetSheetList()
	var sheets [][][]string

	if err != nil {
		fmt.Println(err)
		return [][][]string{}, err
	}

	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for _, sheetName := range sheetsList {
		rows, err2 := f.GetRows(sheetName)
		if err2 != nil {
			fmt.Println(err2)
			return [][][]string{}, err2
		}

		sheets = append(sheets, rows)
	}

	return sheets, nil
}

func (excelFileReader aExcelFileReader) generateModuleClassId(moduleId string, moduleClassName string) string {
	array := strings.Split(moduleClassName, "-")
	arrayLength := len(array)
	return moduleId + "-" + array[arrayLength-3] + "-" + array[arrayLength-2] + "-" + array[arrayLength-1]
}

func (excelFileReader aExcelFileReader) formatCellData(row []string, columnIndex int, typeCell string) string {
	if len(row) == 0 {
		return ""
	}

	cellData := row[columnIndex]
	if typeCell == "room" {
		cellData = regexp.MustCompile(" +").ReplaceAllString(cellData, "")
		cellData = strings.ReplaceAll(cellData, " NCT", "NCT")

		if !strings.Contains(cellData, "-") {
			cellData = "PHTT"
		}
	}

	return cellData
}

func New(filePath string) aExcelFileReader {
	return aExcelFileReader{filePath: filePath}
}
