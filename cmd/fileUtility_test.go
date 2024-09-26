package main

import (
	"bufio"
	"fmt"
	mock_main "gen_work_schedule/mocks"
	"io"
	"os"
	"testing"

	"go.uber.org/mock/gomock"
)

func Test_copyFile(t *testing.T) {
	const TestDirectoryPath = TestDirPath + "/fileUtility"
	const TestSrcFilePath = TestDirectoryPath + "/testsrc.txt"

	// srcを用意する
	if err := os.MkdirAll(TestDirectoryPath, 0755); err != nil {
		t.Errorf("mkdir error: %v", TestDirectoryPath)
	}

	if err := os.WriteFile(
		TestSrcFilePath,
		[]byte("This is test file."),
		0644,
	); err != nil {
		t.Errorf("writing error: %v", err)
	}

	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "正常系",
			args: args{
				src: TestSrcFilePath,
				dst: TestDirectoryPath + "/testdst.txt",
			},
			wantErr: false,
		},
		{
			name:    "sourceが存在しない",
			args:    args{src: TestDirectoryPath + "/xxx"},
			wantErr: true,
		},
		{
			name: "dstが存在しない",
			args: args{
				src: TestSrcFilePath,
				dst: TestDirectoryPath + "/xxx/xxx",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := copyFile(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// コピーできているかチェックする
				srcFile, _ := os.Open(tt.args.src)
				defer srcFile.Close()
				dstFile, _ := os.Open(tt.args.dst)
				defer dstFile.Close()

				srcReader := bufio.NewReader(srcFile)
				dstReader := bufio.NewReader(dstFile)

				for {
					byteSrc, errSrc := srcReader.ReadByte()
					byteDst, errDst := dstReader.ReadByte()

					if errSrc != nil || errDst != nil {
						// 同時にEOFか
						if errSrc == io.EOF && errDst == io.EOF {
							break
						}

						// 片方のみEOFか
						if errSrc == io.EOF || errDst == io.EOF {
							t.Error("src and dst are different in those length")
							break
						}

						// それ以外のエラーか
						t.Errorf("reading error: %v and %v", errSrc, errDst)
						break
					}

					if byteSrc != byteDst {
						srcContents, _ := io.ReadAll(srcFile)
						dstContents, _ := io.ReadAll(dstFile)
						t.Errorf("src and dst are different: src = %v, dst = %v", srcContents, dstContents)
						break
					}
				}
			}
		})
	}
}

func Test_fileExists(t *testing.T) {
	const TestDirectoryPath = TestDirPath + "/fileUtility"
	const TestSrcFilePath = TestDirectoryPath + "/testsrc.txt"
	type args struct {
		path string
	}
	tests := []struct {
		name            string
		args            args
		preconditioning func()
		want            bool
	}{
		{
			name: "正常系",
			args: args{path: TestSrcFilePath},
			preconditioning: func() {
				if err := os.MkdirAll(TestDirectoryPath, 0755); err != nil {
					t.Errorf("mkdir error: %v", err)
				}

				if err := os.WriteFile(
					TestSrcFilePath,
					[]byte("This is test file."),
					0644,
				); err != nil {
					t.Errorf("writing error: %v", err)
				}
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preconditioning()
			if got := fileExists(tt.args.path); got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeCacheDir(t *testing.T) {
	// type args struct {
	// 	cacheDirIF CacheDirInterface
	// }
	tests := []struct {
		name string
		// args    args
		preconditioning func(m *mock_main.MockCacheDirInterface)
		wantErr         bool
	}{
		{
			name: "キャッシュディレクトリ取得エラー",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				m.EXPECT().UserCacheDir().Return("", fmt.Errorf(""))
			},
			wantErr: true,
		},
		{
			name: "既にキャッシュディレクトリが存在する",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				m.EXPECT().UserCacheDir().Return(TestUserCacheDirName, nil)

				// キャッシュディレクトリを用意する
				if !fileExists(TestUserCacheDirName) {
					_ = os.MkdirAll(TestUserCacheDirName, 0755)
				}
			},
			wantErr: false,
		},
		{
			name: "キャッシュディレクトリが存在しない場合は作成する",
			preconditioning: func(m *mock_main.MockCacheDirInterface) {
				m.EXPECT().UserCacheDir().Return(TestUserCacheDirName+"xxx", nil)

				// キャッシュディレクトリが存在する場合は削除しておく
				if fileExists(TestUserCacheDirName + "xxx") {
					os.RemoveAll(TestUserCacheDirName + "xxx")
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock controllerの生成
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// CacheDirInterfaceのモックの生成
			mockCacheDirIF := mock_main.NewMockCacheDirInterface(mockCtrl)
			tt.preconditioning(mockCacheDirIF)

			if err := makeCacheDir(mockCacheDirIF); (err != nil) != tt.wantErr {
				t.Errorf("makeCacheDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
