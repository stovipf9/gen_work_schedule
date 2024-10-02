package main

import (
	"fmt"
	mock_main "gen_work_schedule/mocks"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/xuri/excelize/v2"
	"go.uber.org/mock/gomock"
)

func Test_run(t *testing.T) {
	tests := []struct {
		name            string
		InputReader     io.Reader
		responder       httpmock.Responder
		preconditioning func(ma *MockAppInterfaces)
		wantErr         bool
	}{
		{
			name: "発注情報取り込みでエラー",
			InputReader: io.MultiReader(
				strings.NewReader(strings.Repeat("\n", 3)),
				strings.NewReader("y\n"),
			),
			preconditioning: func(ma *MockAppInterfaces) {
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return("xxx1", nil)
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return("xxx2", nil)
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return("xxx3", nil)
				ma.mz.EXPECT().SelectDirectory().AnyTimes().
					Return(TestDirPath+"/outputs", nil)
				ma.mc.EXPECT().UserCacheDir().AnyTimes().
					Return("", fmt.Errorf("no cache dir"))
			},
			wantErr: true,
		},
		{
			name: "発注情報と名簿の合併でエラー",
			InputReader: io.MultiReader(
				strings.NewReader(strings.Repeat("\n", 3)),
				strings.NewReader("y\n"),
			),
			preconditioning: func(ma *MockAppInterfaces) {
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return(TestDirPath+"/TestOrder.xlsx", nil)
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return("xxx2", nil)
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return("xxx3", nil)
				ma.mz.EXPECT().SelectDirectory().AnyTimes().
					Return(TestDirPath+"/outputs", nil)
				ma.mc.EXPECT().UserCacheDir().AnyTimes().
					Return("", fmt.Errorf("no cache dir"))
			},
			wantErr: true,
		},
		{
			name: "正常系",
			InputReader: io.MultiReader(
				strings.NewReader(strings.Repeat("\n", 3)),
				strings.NewReader("y\n"),
			),
			preconditioning: func(ma *MockAppInterfaces) {
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return(TestDirPath+"/TestOrder.xlsx", nil)
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return(TestDirPath+"/names.csv", nil)
				ma.mz.EXPECT().SelectFile(gomock.Any()).
					Return("xxx3", nil)
				ma.mz.EXPECT().SelectDirectory().AnyTimes().
					Return(TestDirPath+"/outputs", nil)
				ma.mc.EXPECT().UserCacheDir().AnyTimes().
					Return("", fmt.Errorf("no cache dir"))

				// 発注情報を用意する
				_ = os.MkdirAll(TestDirPath, 0755)
				f := excelize.NewFile()
				defer f.Close()

				for colIndex, col := range []struct {
					ColName string
					Value   []string
				}{
					{"会社", []string{"HHH", "HHH"}},
					{"姓", []string{"山田", "山本"}},
					{"発注番号", []string{"123456", "234567"}},
					{"案件名", []string{"２０２４\u3000ＯＯＯ業務", "２０２４\u3000ＸＸＸ業務"}},
					{"PrjCD", []string{"X12Y3456", "Y23Z4567"}},
					{"顧客名", []string{"A社", "B社"}},
					{"担当部署CD", []string{"Z12345", "Z23456"}},
					{"数量", []string{"160", "160"}},
					{"単位", []string{"人時", "人時"}},
					{"作業開始日", []string{"20240901", "20240901"}},
					{"作業終了日", []string{"20240930", "20240930"}},
				} {
					// １行目は空行
					sheetName := f.GetSheetName(0)
					colCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 2)
					_ = f.SetCellStr(sheetName, colCellName, col.ColName)
					for r, v := range col.Value {
						valCellName, _ := excelize.CoordinatesToCellName(colIndex+1, r+3)
						_ = f.SetCellStr(sheetName, valCellName, v)
					}
				}

				_ = f.SaveAs(TestDirPath + "/TestOrder.xlsx")

			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mCtrl := gomock.NewController(t)
			defer mCtrl.Finish()

			// モックの生成
			mockAppIF := MockAppInterfaces{
				mz: mock_main.NewMockZenityInterface(mCtrl),
				mc: mock_main.NewMockCacheDirInterface(mCtrl),
				mi: tt.InputReader,
			}

			// httpモック
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", `^`+regexp.QuoteMeta(HolidaysURL)+`.*$`,
				tt.responder)

			tt.preconditioning(&mockAppIF)

			if err := run(&AppInterfaces{
				ZenityIF:    mockAppIF.mz,
				CacheDirIF:  mockAppIF.mc,
				InputReader: mockAppIF.mi,
			}); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
