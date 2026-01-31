package cmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRosterService_LoadRosterRecords(t *testing.T) {
	const TEST_CSV = "/test.csv"
	tests := []struct {
		name    string // description of this test case
		byteStr string
		path    string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "正常系",
			byteStr: "姓,姓名\n" +
				"山田,山田太郎",
			path:    TEST_CSV,
			want:    map[string]string{"山田": "山田太郎"},
			wantErr: false,
		},
		{
			name:    "ReadFileエラー",
			byteStr: "",
			path:    TEST_CSV + "a",
			want:    make(map[string]string),
			wantErr: true,
		},
		{
			name: "Unmarshalエラー(不正なCSV)",
			byteStr: "姓名\n" +
				"\"",
			path:    TEST_CSV,
			want:    make(map[string]string),
			wantErr: true,
		},
		{
			name: "姓の列が一意でない",
			byteStr: "姓,姓名\n" +
				"山田,山田太郎\n" +
				"山田,山田次郎",
			path:    TEST_CSV,
			want:    make(map[string]string),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFs := afero.NewMemMapFs()
			afero.WriteFile(testFs, TEST_CSV, []byte(tt.byteStr), 0644)
			r := NewRosterService(testFs)
			got, gotErr := r.LoadRosterRecords(tt.path)

			assertErrorOrNoError(t, tt.wantErr, gotErr)

			assert.Equal(t, tt.want, got)
		})
	}
}
