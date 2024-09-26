package main

import (
	"fmt"
	mock_main "gen_work_schedule/mocks"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestCacheDir(t *testing.T) {
	tests := []struct {
		name          string
		prepareMockFn func(mockCacheDirIF *mock_main.MockCacheDirInterface)
		want          string
		wantErr       bool
	}{
		{
			name: "正常系",
			prepareMockFn: func(mockCacheDirIF *mock_main.MockCacheDirInterface) {
				mockCacheDirIF.EXPECT().UserCacheDir().Return(TestUserCacheDirName, nil)
			},
			want:    TestUserCacheDirName + "/GenWorkSchedule/",
			wantErr: false,
		},
		{
			name: "OSキャッシュディレクトリ取得エラー",
			prepareMockFn: func(mockCacheDirIF *mock_main.MockCacheDirInterface) {
				mockCacheDirIF.EXPECT().UserCacheDir().Return("", fmt.Errorf("cache directory does not exist"))
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモックの生成
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			tt.prepareMockFn(mockCacheDirIF)

			got, err := CacheDir(mockCacheDirIF)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CacheDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHolidaysJSONPath(t *testing.T) {
	tests := []struct {
		name          string
		prepareMockFn func(mockCacheDirIF *mock_main.MockCacheDirInterface)
		want          string
		wantErr       bool
	}{
		{
			name: "正常系",
			prepareMockFn: func(mockCacheDirIF *mock_main.MockCacheDirInterface) {
				mockCacheDirIF.EXPECT().UserCacheDir().Times(2).Return(TestUserCacheDirName, nil)
			},
			// want:    cacheDirName + "holidays.json",
			wantErr: false,
		},
		{
			name: "OSキャッシュディレクトリ取得エラー",
			prepareMockFn: func(mockCacheDirIF *mock_main.MockCacheDirInterface) {
				mockCacheDirIF.EXPECT().UserCacheDir().Return("", fmt.Errorf("cache directory does not exist"))
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモックの生成
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			tt.prepareMockFn(mockCacheDirIF)

			// 正常系の場合はCacheDir()の出力を期待結果に用いる
			if !tt.wantErr {
				cacheDirName, _ := CacheDir(mockCacheDirIF)
				tt.want = cacheDirName + "holidays.json"
			}

			got, err := HolidaysJSONPath(mockCacheDirIF)
			if (err != nil) != tt.wantErr {
				t.Errorf("HolidaysJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HolidaysJSONPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathsJSONPath(t *testing.T) {
	tests := []struct {
		name          string
		prepareMockFn func(mockCacheDirIF *mock_main.MockCacheDirInterface)
		want          string
		wantErr       bool
	}{
		{
			name: "正常系",
			prepareMockFn: func(mockCacheDirIF *mock_main.MockCacheDirInterface) {
				mockCacheDirIF.EXPECT().UserCacheDir().Times(2).Return(TestUserCacheDirName, nil)
			},
			// want: cacheDirName + "paths.json",
			wantErr: false,
		},
		{
			name: "OSキャッシュディレクトリ取得エラー",
			prepareMockFn: func(mockCacheDirIF *mock_main.MockCacheDirInterface) {
				mockCacheDirIF.EXPECT().UserCacheDir().Return("", fmt.Errorf("cache directory does not exist"))
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモックの生成
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			tt.prepareMockFn(mockCacheDirIF)

			// 正常系の場合はCacheDir()の出力を期待結果に用いる
			if !tt.wantErr {
				cacheDirName, _ := CacheDir(mockCacheDirIF)
				tt.want = cacheDirName + "paths.json"
			}

			got, err := PathsJSONPath(mockCacheDirIF)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathsJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PathsJSONPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
