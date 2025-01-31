package main

import (
	mock_main "gen_work_schedule/mocks"
	"net/http"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/xuri/excelize/v2"
	"go.uber.org/mock/gomock"
)

func TestFirstSheet(t *testing.T) {
	type args struct {
		// outFile *excelize.File
		orderInfo OrderInfo
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "正常系",
			args: args{orderInfo: OrderInfo{
				ProjectCode:    "X12Y3456",
				DepartmentCode: "Z12345",
				Customer:       "A社",
				Issue:          "２０２４　ＯＯＯ業務",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outFile := excelize.NewFile()
			defer outFile.Close()
			_, _ = outFile.NewSheet("ご使用の前に")

			FirstSheet(outFile, tt.args.orderInfo)

			// PrjCD
			for i, wantCh := range tt.args.orderInfo.ProjectCode {
				cellName, _ := excelize.CoordinatesToCellName(4+i, 18)
				gotCellValue, _ := outFile.GetCellValue(
					"ご使用の前に", cellName)
				if gotCellValue != string(wantCh) {
					t.Errorf("PrjCD: got = %v, want = %v", gotCellValue, string(wantCh))
				}
			}
			// 担当部署CD
			for i, wantch := range tt.args.orderInfo.DepartmentCode {
				cellName, _ := excelize.CoordinatesToCellName(4+i, 19)
				gotCellValue, _ := outFile.GetCellValue(
					"ご使用の前に", cellName,
				)
				if gotCellValue != string(wantch) {
					t.Errorf("担当部署CD: got = %v, want = %v", gotCellValue, string(wantch))
				}
			}
			// 顧客名
			gotCellValue, _ := outFile.GetCellValue(
				"ご使用の前に", "D20",
			)
			if gotCellValue != tt.args.orderInfo.Customer {
				t.Errorf("顧客名: got = %v, want = %v", gotCellValue, tt.args.orderInfo.Customer)
			}
			// 案件名
			gotCellValue, _ = outFile.GetCellValue(
				"ご使用の前に", "D21",
			)
			if gotCellValue != tt.args.orderInfo.Issue {
				t.Errorf("案件名: got = %v, want = %v", gotCellValue, tt.args.orderInfo.Issue)
			}
		})
	}
}

func TestSecondSheet(t *testing.T) {
	const SheetName = "システム要員作業実績報告(委任)"
	type args struct {
		// outFile   *excelize.File
		orderInfo OrderInfo
	}
	tests := []struct {
		name          string
		args          args
		prepareMockFn func(mc *mock_main.MockCacheDirInterface)
		holidaysURL   string
		responder     httpmock.Responder
		workDays      []int
	}{
		{
			name: "正常系",
			args: args{orderInfo: OrderInfo{
				OrderNo:        "123456",
				FullName:       "山田 太郎",
				Issue:          "２０２４　ＯＯＯ業務",
				Representative: "ABC XYZ",
				StartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:        time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			}},
			prepareMockFn: func(mc *mock_main.MockCacheDirInterface) {
				mc.EXPECT().UserCacheDir().AnyTimes().Return(TestUserCacheDirName, nil)
			},
			holidaysURL: HolidaysURL + "/2024/01",
			responder: httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				map[string]interface{}{
					"holidays": []map[string]string{
						{
							"date": "2024-01-01",
							"name": "1st test holiday",
						},
						{
							"date": "2024-01-02",
							"name": "2nd test holiday",
						},
					},
				},
			),
			workDays: []int{
				3, 4, 5,
				8, 9, 10, 11, 12,
				15, 16, 17, 18, 19,
				22, 23, 24, 25, 26,
				29, 30, 31,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモック
			cacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			tt.prepareMockFn(cacheDirIF)

			// httpモック
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", tt.holidaysURL, tt.responder)

			// 編集スプレッドシートの用意
			outFile := excelize.NewFile()
			defer outFile.Close()
			_, _ = outFile.NewSheet(SheetName)

			for row := 20; row <= 50; row++ {
				_ = outFile.SetCellValue(SheetName, "B"+strconv.Itoa(row), row-19)
			}

			SecondSheet(outFile, tt.args.orderInfo, cacheDirIF)

			// 勤務表上部の年月
			gotCellValue, _ := outFile.GetCellValue(SheetName, "G9")
			if gotCellValue != "2024" {
				t.Errorf("年: got = %v, want = %v", gotCellValue, 2024)
			}
			gotCellValue, _ = outFile.GetCellValue(SheetName, "K9")
			if gotCellValue != "1" {
				t.Errorf("月: got = %v, want = %v", gotCellValue, 1)
			}

			// 勤務表上部の発注番号、氏名
			gotCellValue, _ = outFile.GetCellValue(SheetName, "E14")
			if gotCellValue != "S-"+tt.args.orderInfo.OrderNo {
				t.Errorf("発注番号: got = %v, want = %v", gotCellValue, tt.args.orderInfo.OrderNo)
			}
			gotCellValue, _ = outFile.GetCellValue(SheetName, "S16")
			if gotCellValue != tt.args.orderInfo.FullName {
				t.Errorf("姓名: got = %v, want = %v", gotCellValue, tt.args.orderInfo.FullName)
			}

			// 勤務表下部の担当責任者
			gotCellValue, _ = outFile.GetCellValue(SheetName, "S60")
			if gotCellValue != tt.args.orderInfo.Representative {
				t.Errorf("担当責任者: got = %v, want = %v", gotCellValue, tt.args.orderInfo.Representative)
			}

			// 各勤務日
			for row := 20; row <= 50; row++ {
				gotCellValueC, _ := outFile.GetCellValue(SheetName, "C"+strconv.Itoa(row))
				gotCellValueN, _ := outFile.GetCellValue(SheetName, "N"+strconv.Itoa(row))
				if slices.Contains(tt.workDays, row-19) {
					if gotCellValueC != "08:00:00" {
						t.Errorf("row(%v): got = %v, want = %v", row, gotCellValueC, "8:00")
					}
					if gotCellValueN != tt.args.orderInfo.Issue {
						t.Errorf("row(%v): got = %v, want = %v", row, gotCellValueN, tt.args.orderInfo.Issue)
					}
				} else {
					if gotCellValueC != "" {
						t.Errorf("row(%v): got = %v, want = %v", row, gotCellValueC, "")
					}
					if gotCellValueN != "" {
						t.Errorf("row(%v): got = %v, want = %v", row, gotCellValueN, "")
					}
				}

			}
		})
	}
}

func Test_createEachSpreadSheet(t *testing.T) {
	type args struct {
		record    map[string]interface{}
		startDate time.Time
		workDays  []int
		pathOf    PathOf
	}
	tests := []struct {
		name            string
		args            args
		preconditioning func(mc *mock_main.MockCacheDirInterface)
		wantErr         bool
	}{
		{
			name: "正常系",
			args: args{
				record: map[string]interface{}{
					"発注番号":   "123456",
					"会社":     "HHH",
					"姓":      "山田",
					"PrjCD":  "X12Y3456",
					"担当部署CD": "Z12345",
					"顧客名":    "A社",
					"案件名":    "２０２４　ＯＯＯ業務",
					"姓名":     "山田 太郎",
					"作業開始日":  "20240101",
					"作業終了日":  "20240131",
					"担当責任者":  "ABC XYZ",
				},
				startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				workDays:  []int{1, 3},
				pathOf: PathOf{
					TemplateXlsx: TestDirPath + "/template.xlsx",
					OutputDir:    TestDirPath + "/outputs",
				},
			},
			preconditioning: func(mc *mock_main.MockCacheDirInterface) {
				_ = os.MkdirAll(TestDirPath+"/outputs", 0755)
				if !fileExists(TestDirPath + "/template.xlsx") {
					f := excelize.NewFile()
					defer f.Close()

					_ = f.SaveAs(TestDirPath + "/template.xlsx")
				}
				mc.EXPECT().UserCacheDir().AnyTimes().Return(TestUserCacheDirName, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			tt.preconditioning(mockCacheDirIF)

			if err := createEachSpreadSheet(tt.args.record, tt.args.pathOf, mockCacheDirIF); (err != nil) != tt.wantErr {
				t.Errorf("createEachSpreadSheet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
