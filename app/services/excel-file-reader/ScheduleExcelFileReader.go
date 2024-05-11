package excel_file_reader

import (
	"app/utils"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type scheduleExcelFileReader struct {
	aExcelFileReader
	numericalOrderColumnIndex       int
	moduleIdColumnIndex             int
	moduleClassNameColumnIndex      int
	numberOfStudentsColumnIndex     int
	realNumberOfStudentsColumnIndex int
	classTypeColumnIndex            int
	dateColumnIndex                 int
	numberOfWeeksColumnIndex        int
	periods                         []int
	roomColumnIndexes               []int
	academicYearColumnIndex         int

	studySessionId int

	constClassType      map[string]int
	constShiftByPeriods map[string]string
}

func (reader *scheduleExcelFileReader) HandleData() map[string]any {
	sheets, err := reader.ReadAllSheets()

	if err != nil {
		fmt.Println(err)
		return nil
	}

	var moduleClasses = make(map[string]map[string]any)
	var schedules []map[string]string

	fmt.Println(reflect.TypeOf(schedules))
	for _, sheet := range sheets {
		reader.resetColumnIndex()

		moduleId := ""
		moduleClassName := ""
		numberOfStudents := 0
		realNumberOfStudents := 0
		classType := ""
		dateRange := ""
		numberOfWeeks := 0

		isStart := false
		for _, row := range sheet {
			if len(row) == 0 {
				if isStart {
					break
				}

				continue
			}

			if row[reader.numericalOrderColumnIndex] == "STT" {
				for index, cell := range row {
					switch cell {
					case "Mã học phần":
						reader.moduleIdColumnIndex = index

					case "Lớp môn tín chỉ":
						reader.moduleClassNameColumnIndex = index

					case "Số SV DK":
						reader.numberOfStudentsColumnIndex = index

					case "Số SV ĐK":
						reader.realNumberOfStudentsColumnIndex = index

					case "Kiểu học":
						reader.classTypeColumnIndex = index

					case "Thời gian":
						reader.dateColumnIndex = index

					case "Số tuần":
						reader.numberOfWeeksColumnIndex = index

					case "Thứ 2":
						for jndex := 0; jndex < 6; jndex++ {
							reader.periods = append(reader.periods, index+(jndex*2))
							reader.roomColumnIndexes = append(reader.roomColumnIndexes, index+(jndex*2)+1)
						}

					case "Khóa":
						reader.academicYearColumnIndex = index
						break
					}
				}

				continue
			}

			if isStart {
				if len(row) < reader.academicYearColumnIndex-1 {
					break
				}

				moduleId = getSpecificStringIfGivenValueIsEmpty(moduleId, row[reader.moduleIdColumnIndex]).(string)
				moduleClassName = getSpecificStringIfGivenValueIsEmpty(moduleClassName, row[reader.moduleClassNameColumnIndex]).(string)
				numberOfStudents = getSpecificStringIfGivenValueIsEmpty(numberOfStudents, row[reader.numberOfStudentsColumnIndex]).(int)
				realNumberOfStudents = getSpecificStringIfGivenValueIsEmpty(realNumberOfStudents, row[reader.realNumberOfStudentsColumnIndex]).(int)
				classType = getSpecificStringIfGivenValueIsEmpty(classType, row[reader.classTypeColumnIndex]).(string)
				dateRange = getSpecificStringIfGivenValueIsEmpty(dateRange, row[reader.dateColumnIndex]).(string)
				numberOfWeeks = getSpecificStringIfGivenValueIsEmpty(numberOfWeeks, row[reader.numberOfWeeksColumnIndex]).(int)

				reader.createModuleClass(moduleClasses, moduleId, moduleClassName, classType, numberOfStudents, realNumberOfStudents)

				for index, period := range reader.periods {
					if row[period] == "" {
						break
					}

					currentPeriod := row[period]
					currentRoomId := reader.formatCellData(row, reader.roomColumnIndexes[index], "room")
					err := reader.createSchedules(&schedules, moduleId, moduleClassName, dateRange, currentPeriod, currentRoomId, numberOfWeeks, index)
					if err != nil {
						fmt.Println(err)
						return nil
					}
				}
			}

			if reader.moduleIdColumnIndex != -1 && !isStart {
				isStart = true
			}
		}
	}

	return map[string]any{"schedules": schedules, "module_classes": moduleClasses}
}

func (reader *scheduleExcelFileReader) resetColumnIndex() {
	reader.numericalOrderColumnIndex = 1
	reader.moduleIdColumnIndex = -1
	reader.moduleClassNameColumnIndex = -1
	reader.numberOfStudentsColumnIndex = -1
	reader.realNumberOfStudentsColumnIndex = -1
	reader.classTypeColumnIndex = -1
	reader.dateColumnIndex = -1
	reader.periods = []int{}
	reader.roomColumnIndexes = []int{}
	reader.academicYearColumnIndex = -1
}

func getSpecificStringIfGivenValueIsEmpty(currentValue interface{}, value string) interface{} {
	if value == "" {
		return currentValue
	}

	if reflect.TypeOf(currentValue).String() == "int" {
		intValue, _ := strconv.Atoi(value)
		return intValue
	}

	return value
}

func (reader *scheduleExcelFileReader) createModuleClass(moduleClasses map[string]map[string]interface{}, moduleId string,
	moduleClassName string, classType string, numberOfStudents int, realNumberOfStudents int) {

	moduleClassesId := reader.generateModuleClassId(moduleId, moduleClassName)

	isInternational := 0
	if strings.ContainsAny(moduleClassName, "(QT") {
		isInternational = 1
	}

	moduleClasses[moduleClassesId] = map[string]interface{}{
		"id":               moduleClassesId,
		"name":             moduleClassName,
		"number_plan":      numberOfStudents,
		"number_reality":   realNumberOfStudents,
		"type":             reader.constClassType[classType],
		"id_study_session": reader.studySessionId,
		"is_international": isInternational,
		"id_module":        moduleId,
	}
}

func (reader *scheduleExcelFileReader) createSchedules(schedules *[]map[string]string, moduleId string, moduleClassName string, dateRange string, periods string, roomId string, numberOfWeeks int, dayIndexOfWeek int) error {
	moduleClassesId := reader.generateModuleClassId(moduleId, moduleClassName)
	firstDateOfSchedule, err := reader.getFirstDateOfSchedule(dateRange, dayIndexOfWeek)
	if err != nil {
		return err
	}
	shift := reader.getShiftByPeriods(periods)

	for i := 0; i < numberOfWeeks; i++ {
		step := i * 7
		dateString, err2 := utils.CalculateWithDatetimeString(firstDateOfSchedule, "02-01-2006", step)
		if err2 != nil {
			return err2
		}

		*schedules = append(*schedules, map[string]string{
			"id_module_class": moduleClassesId,
			"date":            dateString,
			"shift":           shift,
			"id_room":         roomId,
		})
	}

	return nil
}

func (reader *scheduleExcelFileReader) getFirstDateOfSchedule(dateRange string, dayIndexOfWeek int) (string, error) {
	arrayDateRange := strings.Split(dateRange, "-")
	datetimeString := arrayDateRange[0] + "/" + strings.Split(arrayDateRange[1], "/")[2]
	var err error
	datetimeString, err = utils.ReformatDatetimeString(datetimeString, "02/01/06", "02-01-2006")

	if err != nil {
		return "", err
	}

	var err2 error
	datetimeString, err2 = utils.CalculateWithDatetimeString(datetimeString, "02-01-2006", dayIndexOfWeek)
	if err2 != nil {
		return "", err2
	}

	return datetimeString, nil
}

func (reader *scheduleExcelFileReader) getShiftByPeriods(periods string) string {
	return reader.constShiftByPeriods[regexp.MustCompile(" +").ReplaceAllString(periods, "")]
}

func NewScheduleExcelFileReader(filePath string, studySessionId int) scheduleExcelFileReader {
	reader := scheduleExcelFileReader{}
	reader.aExcelFileReader.filePath = filePath
	reader.studySessionId = studySessionId
	reader.constClassType = map[string]int{
		"LT":  1,
		"BT":  2,
		"TH":  3,
		"DA":  4,
		"BTL": 5,
		"TT":  6,
	}

	reader.constShiftByPeriods = map[string]string{
		"1,2,3":       "1",
		"4,5,6":       "2",
		"7,8,9":       "3",
		"10,11,12":    "4",
		"13,14,15":    "5_1",
		"13,14,15,16": "5_2",
	}

	reader.resetColumnIndex()
	return reader
}
