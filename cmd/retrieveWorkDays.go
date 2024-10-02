package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

// APIのレスポンス内の日付用の型
// RFC2339形式の日時文字列ではないため。
type CustomDate struct {
	time.Time
}

// JSONデコード時の日付パース
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	date, err := time.Parse("2006-01-02", strings.Trim(string(b), "\""))
	if err != nil {
		return err
	}
	cd.Time = date
	return nil
}

type HolidaysJSON struct {
	Holidays []struct {
		Date CustomDate
		Name string
	}
}

func readHolidaysFromJSON(date time.Time, cacheDirIF CacheDirInterface) (HolidaysJSON, error) {
	jsonPath, err := HolidaysJSONPath(cacheDirIF)
	if err != nil {
		return HolidaysJSON{}, err
	}
	// キャッシュを開く
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		return HolidaysJSON{}, err
	}
	defer jsonFile.Close()

	// キャッシュを読む
	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return HolidaysJSON{}, err
	}

	var holidaysJSON HolidaysJSON
	if err := json.NewDecoder(
		strings.NewReader(string(jsonBytes)),
	).Decode(&holidaysJSON); err != nil {
		return HolidaysJSON{}, err
	}

	// キャッシュが当月分のものであるか確認する
	if len(holidaysJSON.Holidays) == 0 {
		return HolidaysJSON{}, fmt.Errorf("JSONが不正")
	}
	dateFromJSON := holidaysJSON.Holidays[0].Date

	if dateFromJSON.Year() != date.Year() || dateFromJSON.Month() != date.Month() {
		return HolidaysJSON{}, fmt.Errorf("当月分の祝日データではない")
	}

	return holidaysJSON, nil
}

func retrieveWorkDays(
	startDate time.Time,
	endDate time.Time,
	cacheDirIF CacheDirInterface,
) ([]int, error) {
	// キャッシュを調べる
	holidaysJSON, err := readHolidaysFromJSON(startDate, cacheDirIF)
	if err != nil {
		// キャッシュからの取得に失敗した場合は
		// https://github.com/shogo82148/holidays-jp
		// によるAPIにより当月の祝日を取得する
		response, err := http.Get(
			fmt.Sprintf(
				HolidaysURL+"/%04d/%02d",
				startDate.Year(), startDate.Month()))
		if err != nil {
			return []int{}, err
		}
		defer response.Body.Close()

		// 取得結果をパースする
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return []int{}, err
		}

		if err := json.NewDecoder(strings.NewReader(string(body))).Decode(&holidaysJSON); err != nil {
			return []int{}, err
		}

		// リクエスト結果をキャッシュする
		jsonPath, err := HolidaysJSONPath(cacheDirIF)
		if err != nil {
			fmt.Println(err)
		} else {
			if err := os.WriteFile(jsonPath, body, 0644); err != nil {
				fmt.Println(err)
			}
		}
	}

	// 勤務日のリストを作成する
	var workDays []int
daysLoop:
	for day := startDate.Day(); day <= endDate.Day(); day++ {
		// 存在しない日付ならばスキップ
		date, err := time.Parse(
			"20060102",
			fmt.Sprintf(
				"%04d%02d%02d",
				startDate.Year(), startDate.Month(), day),
		)
		if err != nil {
			continue
		}

		// 土日ならばスキップ
		if slices.Contains([]time.Weekday{time.Saturday, time.Sunday}, date.Weekday()) {
			continue
		}

		// 祝日ならばスキップ
		for _, holiday := range holidaysJSON.Holidays {
			if holiday.Date.Day() == date.Day() {
				continue daysLoop
			}
		}

		// 勤務日として追加
		workDays = append(workDays, day)
	}

	return workDays, nil
}
