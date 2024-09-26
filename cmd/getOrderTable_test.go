package main

import (
	"os"
	"testing"

	"github.com/go-gota/gota/dataframe"
	"github.com/xuri/excelize/v2"
)

func Test_getOrderTable(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name            string
		args            args
		preconditioning func()
		want            dataframe.DataFrame
		wantErr         bool
	}{
		{
			name: "正常系",
			args: args{path: TestDirPath + "/TestOrder.xlsx"},
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
				} {
					// １行目は空行
					colCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 2)
					valCellName, _ := excelize.CoordinatesToCellName(colIndex+1, 3)
					sheetName := f.GetSheetName(0)
					_ = f.SetCellStr(sheetName, colCellName, col.ColName)
					_ = f.SetCellStr(sheetName, valCellName, col.Value)
				}

				_ = f.SaveAs(TestDirPath + "/TestOrder.xlsx")
			},
			want: dataframe.LoadRecords([][]string{
				{"会社", "姓", "発注番号", "案件名", "PrjCD", "顧客名", "担当部署CD", "数量", "単位", "作業開始日", "作業終了日"},
				{"HHH", "山田", "123456", "２０２４\u3000ＯＯＯ業務", "X12Y3456", "A社", "Z12345", "160", "人時", "20240901", "20240930"}}, dataframe.DetectTypes(false)),
			wantErr: false,
		},
		{
			name:            "スプレッドシートが存在しない",
			args:            args{path: "xxx"},
			preconditioning: func() {},
			want:            dataframe.DataFrame{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preconditioning()
			got, err := getOrderTable(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrderTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareDataframes(got, tt.want) {
				t.Errorf("getOrderTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
