package main

import (
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestFirstSheet(t *testing.T) {
	type args struct {
		// outFile *excelize.File
		record map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "正常系",
			args: args{record: map[string]interface{}{
				"PrjCD":  "X12Y3456",
				"担当部署CD": "Z12345",
				"顧客名":    "A社",
				"案件名":    "２０２４　ＯＯＯ業務",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outFile := excelize.NewFile()
			defer outFile.Close()
			_, _ = outFile.NewSheet("ご使用の前に")

			FirstSheet(outFile, tt.args.record)

			// PrjCD
			for i, wantCh := range tt.args.record["PrjCD"].(string) {
				cellName, _ := excelize.CoordinatesToCellName(4+i, 18)
				gotCellValue, _ := outFile.GetCellValue(
					"ご使用の前に", cellName)
				if gotCellValue != string(wantCh) {
					t.Errorf("PrjCD: got = %v, want = %v", gotCellValue, string(wantCh))
				}
			}
			// 担当部署CD
			for i, wantch := range tt.args.record["担当部署CD"].(string) {
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
			if gotCellValue != tt.args.record["顧客名"].(string) {
				t.Errorf("顧客名: got = %v, want = %v", gotCellValue, tt.args.record["顧客名"])
			}
			// 案件名
			gotCellValue, _ = outFile.GetCellValue(
				"ご使用の前に", "D21",
			)
			if gotCellValue != tt.args.record["案件名"].(string) {
				t.Errorf("案件名: got = %v, want = %v", gotCellValue, tt.args.record["案件名"])
			}
		})
	}
}

func TestSecondSheet(t *testing.T) {
	const SheetName = "システム要員作業実績報告(委任)"
	type args struct {
		// outFile   *excelize.File
		record    map[string]interface{}
		startDate time.Time
		workDays  []int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "正常系",
			args: args{
				record: map[string]interface{}{
					"発注番号": "123456",
					"姓名":   "山田 太郎",
					"案件名":  "２０２４　ＯＯＯ業務",
				},
				startDate: time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC),
				workDays:  []int{1, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outFile := excelize.NewFile()
			defer outFile.Close()
			_, _ = outFile.NewSheet(SheetName)

			for row := 20; row <= 50; row++ {
				_ = outFile.SetCellValue(SheetName, "B"+strconv.Itoa(row), row-19)
			}

			SecondSheet(outFile, tt.args.record, tt.args.startDate, tt.args.workDays)

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
			if gotCellValue != "S-"+tt.args.record["発注番号"].(string) {
				t.Errorf("発注番号: got = %v, want = %v", gotCellValue, tt.args.record["発注番号"])
			}
			gotCellValue, _ = outFile.GetCellValue(SheetName, "S16")
			if gotCellValue != tt.args.record["姓名"].(string) {
				t.Errorf("姓名: got = %v, want = %v", gotCellValue, tt.args.record["姓名"])
			}

			// 各勤務日
			for row := 20; row <= 50; row++ {
				gotCellValueC, _ := outFile.GetCellValue(SheetName, "C"+strconv.Itoa(row))
				gotCellValueN, _ := outFile.GetCellValue(SheetName, "N"+strconv.Itoa(row))
				if slices.Contains(tt.args.workDays, row-19) {
					if gotCellValueC != "08:00:00" {
						t.Errorf("row(%v): got = %v, want = %v", row, gotCellValueC, "8:00")
					}
					if gotCellValueN != tt.args.record["案件名"] {
						t.Errorf("row(%v): got = %v, want = %v", row, gotCellValueN, tt.args.record["案件名"])
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
		preconditioning func()
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
				},
				startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				workDays:  []int{1, 3},
				pathOf: PathOf{
					TemplateXlsx: TestDirPath + "/template.xlsx",
					OutputDir:    TestDirPath + "/outputs",
				},
			},
			preconditioning: func() {
				_ = os.MkdirAll(TestDirPath+"/outputs", 0755)
				if !fileExists(TestDirPath + "/template.xlsx") {
					f := excelize.NewFile()
					defer f.Close()

					_ = f.SaveAs(TestDirPath + "/template.xlsx")
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preconditioning()

			if err := createEachSpreadSheet(tt.args.record, tt.args.startDate, tt.args.workDays, tt.args.pathOf); (err != nil) != tt.wantErr {
				t.Errorf("createEachSpreadSheet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
