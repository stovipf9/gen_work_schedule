package main

import (
	"os"
	"testing"

	"github.com/go-gota/gota/dataframe"
)

func Test_joinFullNames(t *testing.T) {
	const TestCSVPath = TestDirPath + "/names.csv"
	prepareNormalCSV := func() {
		if err := os.MkdirAll(TestDirPath, 0755); err != nil {
			t.Errorf("mkdir error: %v", err)
		}

		if err := os.WriteFile(
			TestCSVPath,
			[]byte("姓,姓名\n山田,山田 太郎"),
			0644,
		); err != nil {
			t.Errorf("writing error: %v", err)
		}
	}
	type args struct {
		namesPath string
		orderDf   dataframe.DataFrame
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
			args: args{
				namesPath: TestCSVPath,
				orderDf: dataframe.LoadRecords([][]string{
					{"姓"},
					{"山田"},
					{"山本"},
				}),
			},
			preconditioning: prepareNormalCSV,
			want: dataframe.LoadMaps([]map[string]interface{}{
				{"姓": "山田", "姓名": "山田 太郎"},
				{"姓": "山本", "姓名": nil},
			}),
			wantErr: false,
		},
		{
			name: "姓の列が無い",
			args: args{
				namesPath: TestCSVPath,
			},
			preconditioning: prepareNormalCSV,
			wantErr:         true,
		},
		{
			name: "CSVなし",
			args: args{
				namesPath: "xxx",
			},
			preconditioning: func() {},
			want:            dataframe.DataFrame{},
			wantErr:         true,
		},
		{
			name: "CSVパスがディレクトリを指している",
			args: args{
				namesPath: TestUserCacheDirName,
			},
			preconditioning: func() {},
			want:            dataframe.DataFrame{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preconditioning()
			got, err := joinFullNames(tt.args.namesPath, tt.args.orderDf)
			if (err != nil) != tt.wantErr {
				t.Errorf("joinFullNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareDataframes(got, tt.want) {
				t.Errorf("joinFullNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
