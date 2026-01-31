package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

var reqlm = NewLockManager(
	func(date time.Time) string { return date.Format("200601") })

type Getter interface {
	Get(url string) (resp *http.Response, err error)
}

type HttpGetter struct{}

func NewHttpGetter() *HttpGetter {
	return &HttpGetter{}
}
func (HttpGetter) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

type HolidaysCacher interface {
	LoadCachedHolidays(date time.Time) (HolidaysJson, error)
	CacheHolidays(holidays []byte) error
}

type WarnLogger interface {
	Warning(message string)
}

type RetrieveWorkdaysService struct {
	getter         Getter
	holidaysCacher HolidaysCacher
	warnLogger     WarnLogger
}

func NewRetrieveWorkdaysService(
	getter Getter,
	holidaysCacher HolidaysCacher,
	warnLogger WarnLogger,
) *RetrieveWorkdaysService {
	return &RetrieveWorkdaysService{
		getter:         getter,
		holidaysCacher: holidaysCacher,
		warnLogger:     warnLogger,
	}
}

func (r *RetrieveWorkdaysService) RetrieveWorkdays(
	startDate time.Time,
	endDate time.Time,
) (workdays []int, error_ error) {

	// キャッシュを調べる
	holidaysJSON, err := r.loadCachedHolidays(startDate)
	if err != nil {
		r.warnLogger.Warning(err.Error())
		// キャッシュからの取得に失敗した場合はリクエスト
		holidaysJSON_, err := r.requestAndCacheHolidays(startDate)
		if err != nil {
			return []int{}, err
		}
		holidaysJSON = holidaysJSON_
	}

	// 勤務日のリストを作成する
	// TODO: 先に祝日抜きで作っておいて、祝日取れなかったかも、みたいな状況にするのが自然では？
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
			// 実質的にはエラー発生しない
			r.warnLogger.Warning(err.Error())
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
		workdays = append(workdays, day)
	}

	return workdays, nil
}

func (r *RetrieveWorkdaysService) requestAndCacheHolidays(
	startDate time.Time,
) (holidays HolidaysJson, _ error) {
	lock := reqlm.GetLock(startDate)
	lock.Lock()
	defer lock.Unlock()

	// https://github.com/shogo82148/holidays-jp
	// によるAPIにより当月の祝日を取得する
	response, err := r.getter.Get(fmt.Sprintf(
		HolidaysURL+"/%04d/%02d",
		startDate.Year(), startDate.Month()))
	if err != nil {
		return holidays, err
	}
	defer response.Body.Close()

	// 取得結果をパース
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return holidays, err
	}
	// リクエスト結果をキャッシュ
	if err := r.holidaysCacher.CacheHolidays(body); err != nil {
		r.warnLogger.Warning(err.Error())
	}

	if err := json.NewDecoder(
		strings.NewReader(string(body))).Decode(&holidays); err != nil {
		return holidays, err
	}

	return holidays, nil
}

func (r *RetrieveWorkdaysService) loadCachedHolidays(date time.Time) (HolidaysJson, error) {
	lock := reqlm.GetLock(date)
	lock.RLock()
	defer lock.RUnlock()

	return r.holidaysCacher.LoadCachedHolidays(date)
}
