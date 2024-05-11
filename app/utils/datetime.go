package utils

import (
	"fmt"
	"time"
)

func ReformatDatetimeString(datetimeString string, oldFormat string, newFormat string) (string, error) {
	datetime, err := time.Parse(oldFormat, datetimeString)
	if err != nil {
		return "", fmt.Errorf("error parsing date: %w", err)
	}

	reformattedDatetimeString := datetime.Format(newFormat)

	return reformattedDatetimeString, nil
}

func CalculateWithDatetimeString(datetimeString string, format string, daysToAdd int) (string, error) {
	datetime, err := time.Parse(format, datetimeString)
	if err != nil {
		return "", err
	}

	return datetime.AddDate(0, 0, daysToAdd).Format(format), nil
}
