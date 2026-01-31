package cmd

import (
	"strings"
	"time"
)

// APIのレスポンス内の日付用の型
// RFC2339形式の日時文字列ではないため。
type HolidayDate struct {
	time.Time
}

// JSONデコード時の日付パース
func (cd *HolidayDate) UnmarshalJSON(b []byte) error {
	date, err := time.Parse("2006-01-02", strings.Trim(string(b), "\""))
	if err != nil {
		return err
	}
	cd.Time = date
	return nil
}

type Holiday struct {
	Date HolidayDate
	Name string
}

type HolidaysJson struct {
	Holidays []Holiday
}
