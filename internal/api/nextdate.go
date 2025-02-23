package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {

	if repeat == "" {
		return "", errors.New("Не указано повторение задачи")
	}

	parseDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("Неправильный формат времени: %s\n", err.Error())
	}
	if repeat == "y" {
		nextDate := parseDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil
	}
	if string(repeat[0]) == "d" && len(repeat) > 2 {
		strDay, _ := strings.CutPrefix(repeat, "d")
		countDay, err := strconv.Atoi(strDay)
		if countDay <= 0 || countDay > 400 || err != nil {
			return "", errors.New("Неправильное количество дней повторения задачи")
		}
		nextDate := parseDate.AddDate(0, 0, countDay)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, countDay)
		}
		return nextDate.Format("20060102"), nil
	} else {
		return "", errors.New("Указан неправильный формат повторения задачи")
	}

}
