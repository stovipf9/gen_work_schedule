package cmd

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewScheduleMakerFactoryWithFs(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fs afero.Fs
	}{
		{
			name: "正常系",
			fs:   afero.NewMemMapFs(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFactory := NewScheduleMakerFactoryWithFs(tt.fs)
			gotEachScheduleMaker := gotFactory(nil, []int{}, "", "", "")

			assert.IsType(t, &EachExcelScheduleMaker{}, gotEachScheduleMaker)
			assert.Equal(t, tt.fs, gotEachScheduleMaker.(*EachExcelScheduleMaker).fs)
		})
	}
}

func TestGenerateService_Generate(t *testing.T) {
	type Mocks struct {
		workdaysRetriever *MockWorkdaysRetriever
		eachScheduleMaker *MockEachScheduleMaker
	}
	tests := []struct {
		name       string // description of this test case
		setupMocks func(*Mocks)
		// Named input parameters for target function.
		orders           []Order
		rosterRecords    map[string]string
		outputDirPath    string
		templateFilePath string
		want             map[string]error
		wantErr          bool
	}{
		{
			name: "正常系",
			setupMocks: func(mocks *Mocks) {
				mocks.workdaysRetriever.EXPECT().RetrieveWorkdays(mock.Anything, mock.Anything).Return([]int{}, nil).Once()
				mocks.eachScheduleMaker.EXPECT().MakeSchedule().Return(nil).Once()
			},
			orders: []Order{
				{
					OrderNo:    "123456",
					Company:    "ABC",
					FamilyName: "山田",
					StartDate:  OrderDate{Time: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
					SequenceNo: 1,
				},
			},
			rosterRecords: map[string]string{
				"山田": "山田 太郎",
			},
			want: map[string]error{
				"S-123456_202409_1_【ABC】山田.xlsx": nil,
			},
			wantErr: false,
		},
		{
			name: "名簿に欠損",
			orders: []Order{
				{FamilyName: "山田"},
			},
			wantErr: true,
		},
		{
			name: "勤務日取得失敗",
			setupMocks: func(mocks *Mocks) {
				mocks.workdaysRetriever.EXPECT().RetrieveWorkdays(mock.Anything, mock.Anything).Return([]int{}, errors.New("RetrieveWorkdays error")).Once()
			},
			orders: []Order{
				{
					OrderNo:    "123456",
					Company:    "ABC",
					FamilyName: "山田",
					StartDate:  OrderDate{Time: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
					SequenceNo: 1,
				},
			},
			rosterRecords: map[string]string{
				"山田": "山田 太郎",
			},
			want: map[string]error{
				"S-123456_202409_1_【ABC】山田.xlsx": errors.New("RetrieveWorkdays error"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := &Mocks{
				workdaysRetriever: NewMockWorkdaysRetriever(t),
				eachScheduleMaker: NewMockEachScheduleMaker(t),
			}
			newEachScheduleMaker := func(
				order *Order,
				workdays []int,
				fullName string,
				outputFilePath string,
				templatePath string,
			) EachScheduleMaker {
				return mocks.eachScheduleMaker
			}
			if tt.setupMocks != nil {
				tt.setupMocks(mocks)
			}
			g := NewGenerateService(mocks.workdaysRetriever, newEachScheduleMaker)
			got, gotErr := g.Generate(tt.orders, tt.rosterRecords, tt.outputDirPath, tt.templateFilePath)

			assertErrorOrNoError(t, tt.wantErr, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
