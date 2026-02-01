package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestRetrieveWorkdaysService_RetrieveWorkdays(t *testing.T) {
	type Mocks struct {
		getter         *MockGetter
		holidaysCacher *MockHolidaysCacher
		warnLogger     *MockWarnLogger
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		setupMocks func(*Mocks)
		// Named input parameters for target function.
		startDate time.Time
		endDate   time.Time
		want      []int
		wantErr   bool
	}{
		{
			name: "正常系",
			setupMocks: func(mocks *Mocks) {
				mocks.holidaysCacher.EXPECT().LoadCachedHolidays(mock.Anything).Return(HolidaysJson{}, errors.New("error in loadCachedHolidays")).Once()
				mocks.warnLogger.EXPECT().Warning(mock.Anything).Once()
				mocks.getter.EXPECT().Get(mock.Anything).Return(
					&http.Response{
						Body: io.NopCloser(strings.NewReader(
							`{"holidays":[{"date":"2024-01-01","name":"1st test holiday"},{"date":"2024-01-02","name":"2nd test holiday"}]}`)),
					},
					nil,
				).Once()
				mocks.holidaysCacher.EXPECT().CacheHolidays(mock.Anything).Return(nil).Once()

			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			want: []int{
				3, 4, 5,
				8, 9, 10, 11, 12,
				15, 16, 17, 18, 19,
				22, 23, 24, 25, 26,
				29, 30, 31,
			},
			wantErr: false,
		},
		{
			name: "祝日取得エラー",
			setupMocks: func(mocks *Mocks) {
				mocks.holidaysCacher.EXPECT().LoadCachedHolidays(mock.Anything).Return(HolidaysJson{}, errors.New("error in loadCachedHolidays")).Once()
				mocks.warnLogger.EXPECT().Warning(mock.Anything).Once()
				mocks.getter.EXPECT().Get(mock.Anything).Return(nil, errors.New("error in Get")).Once()
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			want:      []int{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := &Mocks{
				getter:         NewMockGetter(t),
				holidaysCacher: NewMockHolidaysCacher(t),
				warnLogger:     NewMockWarnLogger(t),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(mocks)
			}
			r := NewRetrieveWorkdaysService(mocks.getter, mocks.holidaysCacher, mocks.warnLogger)

			got, gotErr := r.RetrieveWorkdays(tt.startDate, tt.endDate)

			assertErrorOrNoError(t, tt.wantErr, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRetrieveWorkdaysService_requestAndCacheHolidays(t *testing.T) {
	type Mocks struct {
		getter         *MockGetter
		holidaysCacher *MockHolidaysCacher
		warnLogger     *MockWarnLogger
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		setupMocks func(*Mocks)
		// Named input parameters for target function.
		startDate time.Time
		want      HolidaysJson
		wantErr   bool
	}{
		{
			name: "正常系",
			setupMocks: func(mocks *Mocks) {
				mocks.getter.EXPECT().Get(mock.Anything).Return(
					&http.Response{
						Body: io.NopCloser(strings.NewReader(
							`{"holidays":[{"date":"2024-01-01","name":"1st test holiday"},{"date":"2024-01-02","name":"2nd test holiday"}]}`)),
					},
					nil,
				)
				mocks.holidaysCacher.EXPECT().CacheHolidays(mock.Anything).Return(nil).Once()
			},
			startDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			want: HolidaysJson{
				Holidays: []Holiday{
					{
						Date: HolidayDate{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
						Name: "1st test holiday",
					},
					{
						Date: HolidayDate{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
						Name: "2nd test holiday",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "リクエストエラー",
			setupMocks: func(mocks *Mocks) {
				mocks.getter.EXPECT().Get(mock.Anything).Return(nil, errors.New("request error"))
			},
			startDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantErr:   true,
		},
		{
			name: "リクエスト結果パースエラー",
			setupMocks: func(mocks *Mocks) {
				mocks.getter.EXPECT().Get(mock.Anything).Return(
					&http.Response{
						Body: io.NopCloser(iotest.ErrReader(errors.New("read failed"))),
					},
					nil,
				)
			},
			startDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantErr:   true,
		},
		{
			name: "キャッシュ失敗だがログを残して非nilで返る",
			setupMocks: func(mocks *Mocks) {
				mocks.getter.EXPECT().Get(mock.Anything).Return(
					&http.Response{
						Body: io.NopCloser(strings.NewReader(
							`{"holidays":[{"date":"2024-01-01","name":"1st test holiday"},{"date":"2024-01-02","name":"2nd test holiday"}]}`)),
					},
					nil,
				)
				mocks.holidaysCacher.EXPECT().CacheHolidays(mock.Anything).Return(errors.New("cache error")).Once()
				mocks.warnLogger.EXPECT().Warning(mock.Anything).Once()
			},
			startDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			want: HolidaysJson{
				Holidays: []Holiday{
					{
						Date: HolidayDate{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
						Name: "1st test holiday",
					},
					{
						Date: HolidayDate{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
						Name: "2nd test holiday",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "JSONデコード失敗",
			setupMocks: func(mocks *Mocks) {
				mocks.getter.EXPECT().Get(mock.Anything).Return(
					&http.Response{
						Body: io.NopCloser(strings.NewReader(`{`)),
					},
					nil,
				)
				mocks.holidaysCacher.EXPECT().CacheHolidays(mock.Anything).Return(nil).Once()
			},
			startDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := &Mocks{
				getter:         NewMockGetter(t),
				holidaysCacher: NewMockHolidaysCacher(t),
				warnLogger:     NewMockWarnLogger(t),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(mocks)
			}

			r := NewRetrieveWorkdaysService(mocks.getter, mocks.holidaysCacher, mocks.warnLogger)
			got, gotErr := r.requestAndCacheHolidays(tt.startDate)

			assertErrorOrNoError(t, tt.wantErr, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRetrieveWorkdaysService_loadCachedHolidays(t *testing.T) {
	type MockedLoadResult struct {
		holidaysJson HolidaysJson
		err          error
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		loadResult MockedLoadResult
		want       HolidaysJson
		wantErr    bool
	}{
		{
			name:       "正常系",
			loadResult: MockedLoadResult{HolidaysJson{}, nil},
			want:       HolidaysJson{},
			wantErr:    false,
		},
		{
			name:       "キャッシュロードエラー",
			loadResult: MockedLoadResult{HolidaysJson{}, fmt.Errorf("")},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHolidaysCacher := NewMockHolidaysCacher(t)
			mockHolidaysCacher.EXPECT().LoadCachedHolidays(mock.Anything).
				Return(tt.loadResult.holidaysJson, tt.loadResult.err).Once()
			r := NewRetrieveWorkdaysService(NewMockGetter(t), mockHolidaysCacher, NewMockWarnLogger(t))

			got, gotErr := r.loadCachedHolidays(time.Time{})

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			assert.Equal(t, tt.want, got)
		})
	}
}
