package main

import (
	"errors"
	"fmt"
	mock_main "gen_work_schedule/mocks"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"go.uber.org/mock/gomock"
)

func TestCustomDate_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name         string
		fieldsBefore fields
		args         args
		wantErr      bool
		fieldsAfter  fields
	}{
		{
			name:        "正常系",
			args:        args{b: []byte("2024-09-01")},
			wantErr:     false,
			fieldsAfter: fields{time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:        "二重引用符で囲まれた日付文字列もパースできる",
			args:        args{b: []byte("\"2024-09-01\"")},
			wantErr:     false,
			fieldsAfter: fields{time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:    "空文字はパースエラーとなる",
			args:    args{b: []byte("")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := &CustomDate{
				Time: tt.fieldsBefore.Time,
			}
			if err := cd.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("CustomDate.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cd.Time != tt.fieldsAfter.Time {
				t.Errorf("After CustomDate.UnmarshalJSON() self.Time = %v, want = %v", cd.Time, tt.fieldsAfter.Time)
			}
		})
	}
}

func Test_readHolidaysFromJSON(t *testing.T) {
	var holidaysJSONPath string

	// テスト用のJSONのパスの取得
	func() {
		mCtrl := gomock.NewController(t)
		defer mCtrl.Finish()

		holidaysJSONPath = holidaysJSONPathGetterByMock(mCtrl)
		if holidaysJSONPath == "" {
			t.Errorf("HolidaysJSONPath error")
			return
		}
	}()

	type args struct {
		date time.Time
		// cacheDirIF CacheDirInterface
	}
	tests := []struct {
		name          string
		args          args
		json          string
		wantMockError bool
		want          HolidaysJSON
		wantErr       bool
	}{
		{
			name:          "正常系",
			args:          args{date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			json:          `{"holidays":[{"date": "2024-01-02"},{"date": "2024-01-03"}]}`,
			wantMockError: false,
			want: HolidaysJSON{
				Holidays: []struct {
					Date CustomDate
					Name string
				}{
					{Date: CustomDate{time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)}},
					{Date: CustomDate{time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)}},
				},
			},
			wantErr: false,
		},
		{
			name: "OSキャッシュディレクトリ取得エラー",
			args: args{},
			// precondHolidaysJSON: func() {},
			wantMockError: true,
			want:          HolidaysJSON{},
			wantErr:       true,
		},
		{
			name:          "キャッシュファイルを開くことができない",
			args:          args{},
			wantMockError: false,
			want:          HolidaysJSON{},
			wantErr:       true,
		},
		{
			name:          "JSONファイルのデコード失敗",
			args:          args{},
			json:          "{",
			wantMockError: false,
			want:          HolidaysJSON{},
			wantErr:       true,
		},
		{
			name:          "JSONのデコード失敗 (レコード0件)",
			args:          args{},
			json:          "{}",
			wantMockError: false,
			want:          HolidaysJSON{},
			wantErr:       true,
		},
		{
			name:          "当月のデータではない",
			args:          args{date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			json:          `{"holidays":[{"date": "2000-01-01"}]}`,
			wantMockError: false,
			want:          HolidaysJSON{},
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモックの生成
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			if !tt.wantMockError {
				// os.UserCacheDir()正常想定
				if tt.name == "キャッシュファイルを開くことができない" {
					// 架空のディレクトリを返す
					mockCacheDirIF.EXPECT().UserCacheDir().Return(
						TestUserCacheDirName+"xxx", nil,
					)
				} else {
					mockCacheDirIF.EXPECT().UserCacheDir().Return(
						TestUserCacheDirName, nil,
					)
				}
			} else {
				// os.UserCacheDir()異常想定
				mockCacheDirIF.EXPECT().UserCacheDir().Return(
					"", fmt.Errorf("MockCacheDirInterface error"),
				)
			}

			prepareCacheJSONByMock(t, holidaysJSONPathGetterByMock, tt.json)

			got, err := readHolidaysFromJSON(tt.args.date, mockCacheDirIF)
			if (err != nil) != tt.wantErr {
				t.Errorf("readHolidaysFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readHolidaysFromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_retrieveWorkDays(t *testing.T) {
	type args struct {
		startDate time.Time
		endDate   time.Time
		// cacheDirIF CacheDirInterface
	}
	tests := []struct {
		name        string
		args        args
		cacheJSON   string
		holidaysURL string
		responder   httpmock.Responder
		want        []int
		wantJSON    string
		wantErr     bool
	}{
		{
			name: "正常系",
			args: args{
				startDate: time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 7, 31, 0, 0, 0, 0, time.UTC),
			},
			cacheJSON:   ``,
			holidaysURL: HolidaysURL + "/2024/07",
			responder: httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				map[string]interface{}{
					"holidays": []map[string]string{
						{
							"date": "2024-07-03",
							"name": "1st test holiday",
						},
						{
							"date": "2024-07-25",
							"name": "2nd test holiday",
						},
					},
				},
			),
			want: []int{
				1, 2, 4, 5,
				8, 9, 10, 11, 12,
				15, 16, 17, 18, 19,
				22, 23, 24, 26,
				29, 30, 31,
			},
			wantJSON: `{"holidays":[{"date":"2024-07-03","name":"1st test holiday"},{"date":"2024-07-25","name":"2nd test holiday"}]}`,
			wantErr:  false,
		},
		{
			name: "キャッシュあり",
			args: args{
				startDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			},
			cacheJSON: `{"holidays":[{"date": "2024-06-03"}]}`,
			want: []int{
				4, 5, 6, 7,
				10, 11, 12, 13, 14,
				17, 18, 19, 20, 21,
				24, 25, 26, 27, 28,
			},
			wantErr: false,
		},
		{
			name: "APIリクエストでエラー",
			args: args{
				startDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			},
			cacheJSON:   ``,
			holidaysURL: HolidaysURL + "/2024/06",
			responder: func(_ *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			},
			want:    []int{},
			wantErr: true,
		},
		{
			name: "不完全なレスポンスボディ",
			args: args{
				startDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			},
			cacheJSON:   ``,
			holidaysURL: HolidaysURL + "/2024/06",
			responder: func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(errorResponseReader{}),
				}, nil
			},
			want:    []int{},
			wantErr: true,
		},
		{
			name: "レスポンスボディがJSONとして不正",
			args: args{
				startDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			},
			cacheJSON:   ``,
			holidaysURL: HolidaysURL + "/2024/06",
			responder:   httpmock.NewJsonResponderOrPanic(http.StatusOK, `{`),
			want:        []int{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモックの生成
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			mockCacheDirIF.EXPECT().UserCacheDir().AnyTimes().Return(TestUserCacheDirName, nil)

			prepareCacheJSONByMock(t, holidaysJSONPathGetterByMock, tt.cacheJSON)

			// httpモック
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", tt.holidaysURL, tt.responder)

			got, err := retrieveWorkDays(tt.args.startDate, tt.args.endDate, mockCacheDirIF)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveWorkDays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("retrieveWorkDays() = %v, want %v", got, tt.want)
			}
			if tt.wantJSON != `` {
				// キャッシュされたJSONの内容チェック
				holidaysJSONPath := holidaysJSONPathGetterByMock(mockCtrl)
				jsonFile, _ := os.Open(holidaysJSONPath)
				defer jsonFile.Close()
				jsonBytes, _ := io.ReadAll(jsonFile)
				if string(jsonBytes) != tt.wantJSON {
					t.Errorf("error in writing cache %v, want = %v", string(jsonBytes), tt.wantJSON)
				}
			}
		})
	}
}
