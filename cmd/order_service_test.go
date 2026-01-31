package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

const (
	TestDirPath  = "./tests"
	TestXlsxName = "TestOrder.xlsx"
)

func TestOrderService_LoadOrders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path            string
		preconditioning func()
		want            []Order
		wantErr         bool
	}{
		{
			name: "正常系",
			path: TestDirPath + "/" + TestXlsxName,
			preconditioning: func() {
				_ = os.MkdirAll(TestDirPath, 0755)
				f := excelize.NewFile()
				defer f.Close()

				for colIndex, col := range []struct {
					ColName string
					Value   string
				}{
					{"会社", "HHH"},
					{"姓", "山田"},
					{"発注番号", "123456"},
					{"案件名", "２０２４\u3000ＯＯＯ業務"},
					{"PrjCD", "X12Y3456"},
					{"顧客名", "A社"},
					{"担当部署CD", "Z12345"},
					{"数量", "160"},
					{"単位", "人時"},
					{"作業開始日", "20240901"},
					{"作業終了日", "20240930"},
					{"納入物の提出場所", "開発二部"},
					{"担当責任者", "I. K."},
				} {
					// 1行目は空行
					colCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 2)
					valCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 3)
					sheetName := f.GetSheetName(0)
					_ = f.SetCellStr(sheetName, colCellName, col.ColName)
					_ = f.SetCellStr(sheetName, valCellName, col.Value)
				}

				_ = f.SaveAs(TestDirPath + "/" + TestXlsxName)
			},
			want: []Order{
				{
					Company:        "HHH",
					FamilyName:     "山田",
					OrderNo:        "123456",
					Issue:          "２０２４\u3000ＯＯＯ業務",
					ProjectCode:    "X12Y3456",
					Customer:       "A社",
					DepartmentCode: "Z12345",
					Amount:         160,
					Unit:           "人時",
					StartDate:      OrderDate{Time: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
					EndDate:        OrderDate{Time: time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC)},
					Department:     "開発二部",
					Representative: "I. K.",
				},
			},
			wantErr: false,
		},
		{
			name:            "スプレッドシートが存在しない",
			path:            "xxx",
			preconditioning: func() {},
			want:            nil,
			wantErr:         true,
		},
		{
			name: "正常系 (シート下方に空行あり)",
			path: TestDirPath + "/" + TestXlsxName,
			preconditioning: func() {
				_ = os.MkdirAll(TestDirPath, 0755)
				f := excelize.NewFile()
				defer f.Close()

				sheetName := f.GetSheetName(0)
				for colIndex, col := range []struct {
					ColName string
					Value   string
				}{
					{"会社", "HHH"},
					{"姓", "山田"},
					{"発注番号", "123456"},
					{"案件名", "２０２４\u3000ＯＯＯ業務"},
					{"PrjCD", "X12Y3456"},
					{"顧客名", "A社"},
					{"担当部署CD", "Z12345"},
					{"数量", "160"},
					{"単位", "人時"},
					{"作業開始日", "20240901"},
					{"作業終了日", "20240930"},
					{"納入物の提出場所", "開発二部"},
					{"担当責任者", "I. K."},
				} {
					// 1行目は空行
					colCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 2)
					valCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 3)
					_ = f.SetCellStr(sheetName, colCellName, col.ColName)
					_ = f.SetCellStr(sheetName, valCellName, col.Value)
				}
				// 空行を挟んで
				jumpedCellName, _ := excelize.CoordinatesToCellName(1, 10)
				_ = f.SetCellStr(sheetName, jumpedCellName, "dummy")

				_ = f.SaveAs(TestDirPath + "/" + TestXlsxName)
			},
			want: []Order{
				{
					Company:        "HHH",
					FamilyName:     "山田",
					OrderNo:        "123456",
					Issue:          "２０２４\u3000ＯＯＯ業務",
					ProjectCode:    "X12Y3456",
					Customer:       "A社",
					DepartmentCode: "Z12345",
					Amount:         160,
					Unit:           "人時",
					StartDate:      OrderDate{Time: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)},
					EndDate:        OrderDate{Time: time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC)},
					Department:     "開発二部",
					Representative: "I. K.",
				},
			},
			wantErr: false,
		},
		{
			name: "空Excel",
			path: TestDirPath + "/" + TestXlsxName,
			preconditioning: func() {
				_ = os.MkdirAll(TestDirPath, 0755)
				f := excelize.NewFile()
				defer f.Close()

				_ = f.SaveAs(TestDirPath + "/" + TestXlsxName)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "デコードエラー (日時の形式不正)",
			path: TestDirPath + "/" + TestXlsxName,
			preconditioning: func() {
				_ = os.MkdirAll(TestDirPath, 0755)
				f := excelize.NewFile()
				defer f.Close()

				for colIndex, col := range []struct {
					ColName string
					Value   string
				}{
					{"会社", "HHH"},
					{"姓", "山田"},
					{"発注番号", "123456"},
					{"案件名", "２０２４\u3000ＯＯＯ業務"},
					{"PrjCD", "X12Y3456"},
					{"顧客名", "A社"},
					{"担当部署CD", "Z12345"},
					{"数量", "160"},
					{"単位", "人時"},
					{"作業開始日", "2024/09/01"},
					{"作業終了日", "2024/09/30"},
					{"納入物の提出場所", "開発二部"},
					{"担当責任者", "I. K."},
				} {
					// 1行目は空行
					colCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 2)
					valCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 3)
					sheetName := f.GetSheetName(0)
					_ = f.SetCellStr(sheetName, colCellName, col.ColName)
					_ = f.SetCellStr(sheetName, valCellName, col.Value)
				}

				_ = f.SaveAs(TestDirPath + "/" + TestXlsxName)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preconditioning()
			o := NewOrderService()
			got, gotErr := o.LoadOrders(tt.path)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			assert.Equal(t, tt.want, got)
		})
	}
}
