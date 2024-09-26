package main

import (
	"fmt"
	mock_main "gen_work_schedule/mocks"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
)

func Test_confirm(t *testing.T) {
	type args struct {
		reader io.Reader
		prompt string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "yes",
			args: args{
				reader: strings.NewReader("y\n"),
				prompt: "テストプロンプト",
			},
			want: true,
		},
		{
			name: "no",
			args: args{
				reader: strings.NewReader("n\n"),
				prompt: "テストプロンプト",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := confirm(tt.args.reader, tt.args.prompt); got != tt.want {
				t.Errorf("confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readPathsFromJSON(t *testing.T) {
	prepareNormalMock := func(m *mock_main.MockCacheDirInterface) {
		m.EXPECT().UserCacheDir().AnyTimes().Return(TestUserCacheDirName, nil)
	}
	prepareJSON := func(m *mock_main.MockCacheDirInterface, json string) {
		pathsJSONPath, _ := PathsJSONPath(m)
		f, _ := os.OpenFile(
			pathsJSONPath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0644,
		)
		defer f.Close()

		if _, err := f.WriteString(json); err != nil {
			t.Errorf("erro in creating json: %v\n", err)
		}
	}

	type args struct {
		// cacheDirIF CacheDirInterface
	}
	tests := []struct {
		name            string
		args            args
		preconditioning func(m *mock_main.MockCacheDirInterface)
		want            PathOf
		wantErr         bool
	}{
		{
			name: "正常系。ただし記載されたパスにファイルが存在しない場合は値を捨てる",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				prepareNormalMock(m)
				prepareJSON(m,
					`{
					"OrderXlsx":"xxx",
					"NamesCSV": "xxx",
					"TemplateXlsx": "xxx",
					"OutputDir": "xxx"
					}`,
				)
			},
			want:    PathOf{OutputDir: "xxx"},
			wantErr: false,
		},
		{
			name: "キャッシュディレクトリ取得エラー",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				m.EXPECT().UserCacheDir().Return("", fmt.Errorf("cache directory error"))
			},
			want:    PathOf{},
			wantErr: true,
		},
		{
			name: "キャッシュファイルが存在しない",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				m.EXPECT().UserCacheDir().Return(TestUserCacheDirName+"xxx", nil)
			},
			want:    PathOf{},
			wantErr: true,
		},
		{
			name: "JSONファイルの内容が不正",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				prepareNormalMock(m)
				prepareJSON(m, "{")
			},
			want:    PathOf{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモック
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)

			tt.preconditioning(mockCacheDirIF)

			got, err := readPathsFromJSON(mockCacheDirIF)
			if (err != nil) != tt.wantErr {
				t.Errorf("readPathsFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readPathsFromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inputPaths(t *testing.T) {
	const TestJSONPath = TestUserCacheDirName + "/paths.json"

	prepareNormalMock := func(ma *MockAppInterfaces) {
		ma.mz.EXPECT().SelectFile(gomock.Any()).Times(3).
			Return("xxx", nil)
		ma.mz.EXPECT().SelectDirectory().
			Return("yyy", nil)
		ma.mc.EXPECT().UserCacheDir().Times(2).
			Return(TestUserCacheDirName, nil)
	}
	prepareJSON := func(json string) {
		f, _ := os.OpenFile(
			TestJSONPath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0644,
		)
		defer f.Close()

		if _, err := f.WriteString(json); err != nil {
			t.Errorf("error in creating json: %v\n", err)
		}
	}

	type args struct {
		// appIF *AppInterfaces
	}
	tests := []struct {
		name            string
		args            args
		InputReader     io.Reader
		preconditioning func(ma *MockAppInterfaces)
		want            PathOf
		wantErr         bool
	}{
		{
			name: "正常系(ディレクトリパスのみキャッシュ採用)",
			InputReader: io.MultiReader(
				// ファイル3つについては選択させる
				strings.NewReader(strings.Repeat("\n", 3)),
				// ディレクトリについてはキャッシュを採用
				strings.NewReader("y\n"),
			),
			preconditioning: func(ma *MockAppInterfaces) {
				prepareJSON(
					`{
						"OrderXlsx":"",
						"NamesCSV": "",
						"TemplateXlsx": "",
						"OutputDir": "yyy"
					}`,
				)
				prepareNormalMock(ma)
			},
			want: PathOf{
				OrderXlsx:    "xxx",
				NamesCSV:     "xxx",
				TemplateXlsx: "xxx",
				OutputDir:    "yyy",
			},
			wantErr: false,
		},
		{
			name: "正常系(ディレクトリパスキャッシュ不採用)",
			InputReader: io.MultiReader(
				// ファイル3つについては選択させる
				strings.NewReader(strings.Repeat("\n", 3)),
				// ディレクトリについてはキャッシュ不採用
				strings.NewReader("n\n"),
				strings.NewReader("\n"),
			),
			preconditioning: func(ma *MockAppInterfaces) {
				prepareJSON(
					`{
						"OrderXlsx":"",
						"NamesCSV": "",
						"TemplateXlsx": "",
						"OutputDir": "yyy"
					}`,
				)
				ma.mz.EXPECT().SelectFile(gomock.Any()).Times(3).
					Return("xxx", nil)
				ma.mz.EXPECT().SelectDirectory().
					Return("zzz", nil)
				ma.mc.EXPECT().UserCacheDir().Times(2).
					Return(TestUserCacheDirName, nil)
			},
			want: PathOf{
				OrderXlsx:    "xxx",
				NamesCSV:     "xxx",
				TemplateXlsx: "xxx",
				OutputDir:    "zzz",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// 各モックの生成
			mockAppIF := MockAppInterfaces{
				mz: mock_main.NewMockZenityInterface(mockCtrl),
				mc: mock_main.NewMockCacheDirInterface(mockCtrl),
				mi: tt.InputReader,
			}

			tt.preconditioning(&mockAppIF)

			got, err := inputPaths(
				&AppInterfaces{
					ZenityIF:    mockAppIF.mz,
					CacheDirIF:  mockAppIF.mc,
					InputReader: mockAppIF.mi,
				},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("inputPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inputPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}
