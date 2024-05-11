package excel_file_reader

type iExcelFileReader interface {
	ReadFirstSheet() ([][]string, error)
}
