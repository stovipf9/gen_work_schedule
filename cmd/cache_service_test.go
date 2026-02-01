package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func assertErrorOrNoError(t *testing.T, wantErr bool, err error) {
	if wantErr {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

func assertJsonEq(t *testing.T, fs afero.Fs, path string, want string) {
	bytes, err := afero.ReadFile(fs, path)
	assert.NoError(t, err)
	assert.JSONEq(t, want, string(bytes))
}

func TestCacheService_LoadCachedPaths(t *testing.T) {
	const DEDICATED_CACHE_DIR_NAME = "temp"
	tests := []struct {
		name        string // description of this test case
		prepareFile func(afero.Fs)
		want        map[PathKey]string
		wantErr     bool
	}{
		{
			name: "正常系",
			prepareFile: func(fs afero.Fs) {
				afero.WriteFile(fs, DEDICATED_CACHE_DIR_NAME+"/"+PATHS_JSON,
					[]byte(`{"OrderXlsx":"order.xlsx"}`), 0644)
			},
			want: map[PathKey]string{
				OrderXlsx: "order.xlsx",
			},
			wantErr: false,
		},
		{
			name:    "キャッシュファイルが存在しない",
			want:    map[PathKey]string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFs := afero.NewMemMapFs()
			c := NewCacheService(testFs, DEDICATED_CACHE_DIR_NAME)
			if tt.prepareFile != nil {
				tt.prepareFile(testFs)
			}
			got, gotErr := c.LoadCachedPaths()

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCacheService_CachePaths(t *testing.T) {
	const DEDICATED_CACHE_DIR_NAME = "temp"
	tests := []struct {
		name string // description of this test case
		fs   afero.Fs
		// Named input parameters for target function.
		paths        map[PathKey]string
		wantJsonPath string
		assert       func(fs afero.Fs, path string)
		wantErr      bool
	}{
		{
			name: "正常系",
			fs:   afero.NewMemMapFs(),
			paths: map[PathKey]string{
				OrderXlsx: "order.xlsx",
			},
			wantJsonPath: DEDICATED_CACHE_DIR_NAME + "/" + PATHS_JSON,
			assert: func(fs afero.Fs, path string) {
				want := []byte(`{"OrderXlsx":"order.xlsx"}`)
				assertJsonEq(t, fs, path, string(want))
			},
			wantErr: false,
		},
		{
			name: "キャッシュファイル書き込みエラー",
			fs:   afero.NewReadOnlyFs(afero.NewMemMapFs()),
			paths: map[PathKey]string{
				OrderXlsx: "order.xlsx",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCacheService(tt.fs, DEDICATED_CACHE_DIR_NAME)
			gotErr := c.CachePaths(tt.paths)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			if tt.assert != nil {
				tt.assert(tt.fs, tt.wantJsonPath)
			}
		})
	}
}

func TestCacheService_dateToFileName(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		date time.Time
		want string
	}{
		{
			name: "正常系",
			date: time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
			want: "holidays_202409.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CacheService{}.dateToFileName(tt.date)
			if got != tt.want {
				t.Errorf("dateToFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheService_LoadCachedHolidays(t *testing.T) {
	const DEDICATED_CACHE_DIR_NAME = "temp"
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		prepareFile func(afero.Fs)
		// Named input parameters for target function.
		date    time.Time
		want    HolidaysJson
		wantErr bool
	}{
		{
			name: "正常系",
			prepareFile: func(fs afero.Fs) {
				afero.WriteFile(fs, DEDICATED_CACHE_DIR_NAME+"/"+fmt.Sprintf(HOLIDAYS_JSON, "202101"),
					[]byte(`{"holidays":[{"date":"2021-01-01","name":"元日"}]}`), 0644)
			},
			date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			want: HolidaysJson{
				Holidays: []Holiday{
					{
						Date: HolidayDate{Time: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
						Name: "元日",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "ファイルが存在しない",
			date:    time.Now(),
			wantErr: true,
		},
		{
			name: "祝日が無い",
			prepareFile: func(fs afero.Fs) {
				afero.WriteFile(fs, DEDICATED_CACHE_DIR_NAME+"/"+fmt.Sprintf(HOLIDAYS_JSON, "202101"),
					[]byte(`{"holidays":[]}`), 0644)
			},
			date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			// want: []Holiday{},
			wantErr: true,
		},
		{
			name: "求める月の祝日ではない",
			prepareFile: func(fs afero.Fs) {
				afero.WriteFile(fs, DEDICATED_CACHE_DIR_NAME+"/"+fmt.Sprintf(HOLIDAYS_JSON, "202101"),
					[]byte(`{"holidays":[{"date":"2022-01-01","name":"元日"}]}`), 0644)
			},
			date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			// want:
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFs := afero.NewMemMapFs()
			c := NewCacheService(testFs, DEDICATED_CACHE_DIR_NAME)
			if tt.prepareFile != nil {
				tt.prepareFile(testFs)
			}
			got, gotErr := c.LoadCachedHolidays(tt.date)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCacheService_CacheHolidays(t *testing.T) {
	const DEDICATED_CACHE_DIR_NAME = "temp"
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		dedicatedCacheDirName string
		// Named input parameters for target function.
		rawHolidays  []byte
		assert       func(fs afero.Fs, path string, want []byte)
		wantJsonPath string
		wantErr      bool
	}{
		{
			name:        "正常系",
			rawHolidays: []byte(`{"holidays":[{"date":"2021-01-01","name":"元日"}]}`),
			wantJsonPath: DEDICATED_CACHE_DIR_NAME + "/" +
				fmt.Sprintf(HOLIDAYS_JSON, "202101"),
			assert: func(fs afero.Fs, path string, want []byte) {
				assertJsonEq(t, fs, path, string(want))
			},
			wantErr: false,
		},
		{
			name:        "データパース不能",
			rawHolidays: []byte("{"),
			wantErr:     true,
		},
		{
			name:        "祝日がゼロ",
			rawHolidays: []byte(`{"holidays":[]}`),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFs := afero.NewMemMapFs()
			c := NewCacheService(testFs, DEDICATED_CACHE_DIR_NAME)
			gotErr := c.CacheHolidays(tt.rawHolidays)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			if tt.assert != nil {
				tt.assert(testFs, tt.wantJsonPath, tt.rawHolidays)
			}
		})
	}
}

// func TestCacheService_LoadCachedWindowPosSize(t *testing.T) {
// 	tests := []struct {
// 		name                  string // description of this test case
// 		preconditioning       func(afero.Fs, string)
// 		cacheDirfs            afero.Fs
// 		dedicatedCacheDirName string
// 		want                  WindowPosSize
// 		wantErr               bool
// 	}{
// 		{
// 			name:                  "正常系",
// 			cacheDirfs:            afero.NewMemMapFs(),
// 			dedicatedCacheDirName: "temp",
// 			preconditioning: func(fs afero.Fs, path string) {
// 				afero.WriteFile(afero.NewBasePathFs(fs, path), WINDOW_POS_SIZE_JSON, []byte("{}"), 0644)
// 			},
// 			want:    WindowPosSize{},
// 			wantErr: false,
// 		},
// 		{
// 			name:                  "キャッシュファイルopenエラー",
// 			cacheDirfs:            afero.NewMemMapFs(),
// 			dedicatedCacheDirName: "temp",
// 			wantErr:               true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			testFs := afero.NewMemMapFs()
// 			c := NewCacheService(testFs, tt.dedicatedCacheDirName)
// 			if tt.preconditioning != nil {
// 				tt.preconditioning(testFs, tt.dedicatedCacheDirName)
// 			}
// 			got, gotErr := c.LoadCachedWindowPosSize()

// 			assertErrorOrNoError(t, tt.wantErr, gotErr)

// 			assert.Equal(t, tt.want, got)
// 		})
// 	}
// }

// func TestCacheService_CacheWindowPosSize(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		fs   afero.Fs
// 		// Named input parameters for target function.
// 		windowRect WindowPosSize
// 		wantErr    bool
// 	}{
// 		{
// 			name: "正常系",
// 			fs:   afero.NewMemMapFs(),
// 			windowRect: WindowPosSize{
// 				Pos:  Pos{1, 2},
// 				Size: Size{3, 4},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:       "書き込みエラー",
// 			fs:         afero.NewReadOnlyFs(afero.NewMemMapFs()),
// 			windowRect: WindowPosSize{},
// 			wantErr:    true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			const DIR_NAME = "gen_work_schedule"
// 			c := NewCacheService(tt.fs, DIR_NAME)
// 			gotErr := c.CacheWindowPosSize(tt.windowRect)

// 			assertErrorOrNoError(t, tt.wantErr, gotErr)
// 		})
// 	}
// }

func Test_cacheData(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fs       afero.Fs
		data     any
		jsonPath string
		assert   func(afero.Fs, string)
		wantErr  bool
	}{
		{
			name:     "正常系",
			fs:       afero.NewMemMapFs(),
			data:     time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
			jsonPath: "/temp/a.json",
			assert: func(fs afero.Fs, path string) {
				want := []byte(`"2024-09-30T00:00:00Z"`)
				assertJsonEq(t, fs, path, string(want))

			},
			wantErr: false,
		},
		{
			name:    "デシリアライズエラー",
			fs:      afero.NewMemMapFs(),
			data:    func() {}, // 関数型はデシリアライズできない
			wantErr: true,
		},
		{
			name:     "キャッシュファイル書き込みエラー",
			fs:       afero.NewReadOnlyFs(afero.NewMemMapFs()), // 読み取り専用
			data:     "test",
			jsonPath: "/temp/a.json",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := CacheData(tt.fs, tt.data, tt.jsonPath)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			if tt.assert != nil {
				tt.assert(tt.fs, tt.jsonPath)
			}
		})
	}
}

func Test_loadCachedData(t *testing.T) {
	type TestCase[U any] struct {
		name         string // description of this test case
		fs           afero.Fs
		jsonPath     string
		prepareCache func(afero.Fs, string)
		want         U
		wantErr      bool
	}
	tests := []TestCase[time.Time]{
		{
			name:     "正常系",
			jsonPath: "/temp/a.json",
			prepareCache: func(fs afero.Fs, path string) {
				afero.WriteFile(fs, path, []byte("\"2024-09-30T00:00:00Z\""), 0644)
			},
			want:    time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:     "キャッシュファイルが存在しない",
			jsonPath: "invalidPath",
			wantErr:  true,
		},
		{
			name:     "シリアライズエラー",
			jsonPath: "/temp/a.json",
			prepareCache: func(fs afero.Fs, path string) {
				afero.WriteFile(fs, path, []byte("{"), 0644)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fs == nil {
				tt.fs = afero.NewMemMapFs()
			}
			if tt.prepareCache != nil {
				tt.prepareCache(tt.fs, tt.jsonPath)
			}
			got, gotErr := LoadCachedData[time.Time](tt.fs, tt.jsonPath)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			assert.Equal(t, tt.want, got)
		})
	}
}
