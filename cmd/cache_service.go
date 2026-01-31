package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

var filelm = NewLockManager(func(key string) string { return key })

const (
	PATHS_JSON    = "paths.json"
	HOLIDAYS_JSON = "holidays_%s.json"
)

type CacheService struct {
	cacheDirFs afero.Fs
}

func NewCacheService(
	cacheDirfs afero.Fs,
	dedicatedCacheDirName string,
) *CacheService {
	return &CacheService{cacheDirFs: afero.NewBasePathFs(cacheDirfs, dedicatedCacheDirName)}
}

func (c CacheService) CacheDirFs() afero.Fs {
	return c.cacheDirFs
}

func (c CacheService) LoadCachedPaths() (map[PathKey]string, error) {
	paths, err := LoadCachedData[map[PathKey]string](
		c.cacheDirFs,
		PATHS_JSON,
	)
	if paths == nil || err != nil {
		return make(map[PathKey]string), err
	}

	return paths, nil
}

func (c CacheService) CachePaths(paths map[PathKey]string) error {
	return CacheData(
		c.cacheDirFs,
		paths,
		PATHS_JSON,
	)
}

func (CacheService) dateToFileName(date time.Time) string {
	return fmt.Sprintf(HOLIDAYS_JSON, date.Format("200601"))
}

func (c CacheService) LoadCachedHolidays(date time.Time) (
	holidays HolidaysJson,
	_ error,
) {
	holidays, err := LoadCachedData[HolidaysJson](
		c.cacheDirFs,
		CacheService{}.dateToFileName(date),
	)
	if err != nil {
		return holidays, err
	}

	// キャッシュが当月分のものであるか確認する
	if len(holidays.Holidays) == 0 {
		// 祝日0日
		return HolidaysJson{}, fmt.Errorf("JSONが不正")
	}
	dateFromJson := holidays.Holidays[0].Date
	if dateFromJson.Year() != date.Year() ||
		dateFromJson.Month() != date.Month() {
		return HolidaysJson{}, fmt.Errorf("当月分の祝日データではない")
	}

	return holidays, nil
}

func (c CacheService) CacheHolidays(rawHolidays []byte) error {

	var holidays HolidaysJson
	if err := json.Unmarshal(rawHolidays, &holidays); err != nil {
		return err
	}
	if len(holidays.Holidays) == 0 {
		return fmt.Errorf("No holidays in %v", rawHolidays)
	}

	return CacheData(
		c.cacheDirFs,
		rawHolidays,
		CacheService{}.dateToFileName(holidays.Holidays[0].Date.Time),
	)
}

// func (c CacheService) LoadCachedWindowPosSize() (WindowPosSize, error) {
// 	return loadCachedData[WindowPosSize](
// 		c.cacheDirFs,
// 		WINDOW_POS_SIZE_JSON,
// 	)
// }

// func (c CacheService) CacheWindowPosSize(windowRect WindowPosSize) error {
// 	return cacheData(
// 		c.cacheDirFs,
// 		windowRect,
// 		WINDOW_POS_SIZE_JSON,
// 	)
// }

func CacheData[T any](
	fs afero.Fs,
	data T,
	jsonPath string,
) error {
	// キャッシュファイルを読み書きロック
	lock := filelm.GetLock(jsonPath)
	lock.Lock()
	defer lock.Unlock()

	// デシリアライズ
	// 既にデータがバイト列なら不要
	jsonBytes, ok := any(data).([]byte)
	if !ok {
		var err error
		jsonBytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	if err := fs.MkdirAll(filepath.Dir(jsonPath), 0755); err != nil {
		return err
	}

	// 書き込み
	if err := afero.WriteFile(fs, jsonPath, jsonBytes, 0666); err != nil {
		return err
	}

	return nil
}

func LoadCachedData[T any](
	fs afero.Fs,
	jsonPath string,
) (cachedData T, _ error) {
	lock := filelm.GetLock(jsonPath)
	lock.RLock()
	defer lock.RUnlock()

	// キャッシュを読み込む
	jsonBytes, err := afero.ReadFile(fs, jsonPath)
	if err != nil {
		return cachedData, err
	}

	// シリアライズ
	if err := json.NewDecoder(
		strings.NewReader(string(jsonBytes)),
	).Decode(&cachedData); err != nil {
		return cachedData, err
	}

	return cachedData, nil
}
