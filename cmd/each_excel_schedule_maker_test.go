package cmd

import (
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestEachExcelScheduleMaker_MakeSchedule(t *testing.T) {
	const TEMPLATE_XLSX = "template.xlsx"
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		fs              afero.Fs
		order           *Order
		workdays        []int
		fullName        string
		outputFilePath  string
		templatePath    string
		prepareTemplate func(fs afero.Fs, templatePath string)
		wantErr         bool
	}{
		{
			name:           "正常系",
			fs:             afero.NewOsFs(),
			order:          &Order{},
			outputFilePath: filepath.Join(t.TempDir(), "out.xlsx"),
			templatePath:   filepath.Join(t.TempDir(), TEMPLATE_XLSX),
			prepareTemplate: func(fs afero.Fs, templatePath string) {
				f := excelize.NewFile()
				// 「ご使用の前に」シート作成
				if _, err := f.NewSheet(BeforeUseSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 「システム要員作業実績報告(委任)」シート作成
				if _, err := f.NewSheet(OutsourceSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 勤務日行頭埋め
				for row := 20; row <= 50; row++ {
					if err := f.SetCellValue(OutsourceSheetName, "B"+strconv.Itoa(row), row-19); err != nil {
						assert.FailNow(t, err.Error())
					}
				}
				// 「御見積書 (委任)」シート作成
				if _, err := f.NewSheet(QuotationDelegationSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				if err := f.SaveAs(templatePath); err != nil {
					assert.FailNow(t, err.Error())
				}
			},
			wantErr: false,
		},
		{
			name:         "テンプレートファイル無し",
			fs:           afero.NewMemMapFs(),
			templatePath: "non_existent_file",
			wantErr:      true,
		},
		{
			name:    "出力名不正により出力ファイルが無い",
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		{
			name:           "「ご使用の前に」シート無し",
			fs:             afero.NewOsFs(),
			order:          &Order{},
			outputFilePath: filepath.Join(t.TempDir(), "out.xlsx"),
			templatePath:   filepath.Join(t.TempDir(), TEMPLATE_XLSX),
			prepareTemplate: func(fs afero.Fs, templatePath string) {
				f := excelize.NewFile()
				// 「ご使用の前に」シートを作らない
				if err := f.SaveAs(templatePath); err != nil {
					assert.FailNow(t, err.Error())
				}
			},
			wantErr: true,
		},
		{
			name:           "「システム要員作業実績報告(委任)」シート無し",
			fs:             afero.NewOsFs(),
			order:          &Order{},
			outputFilePath: filepath.Join(t.TempDir(), "out.xlsx"),
			templatePath:   filepath.Join(t.TempDir(), TEMPLATE_XLSX),
			prepareTemplate: func(fs afero.Fs, templatePath string) {
				f := excelize.NewFile()
				// 「ご使用の前に」シート作成
				if _, err := f.NewSheet(BeforeUseSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 「システム要員作業実績報告(委任)」シートを作らない
				if err := f.SaveAs(templatePath); err != nil {
					assert.FailNow(t, err.Error())
				}
			},
			wantErr: true,
		},
		{
			name:           "「御見積書 (委任)」シート無し",
			fs:             afero.NewOsFs(),
			order:          &Order{},
			outputFilePath: filepath.Join(t.TempDir(), "out.xlsx"),
			templatePath:   filepath.Join(t.TempDir(), TEMPLATE_XLSX),
			prepareTemplate: func(fs afero.Fs, templatePath string) {
				f := excelize.NewFile()
				// 「ご使用の前に」シート作成
				if _, err := f.NewSheet(BeforeUseSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 「システム要員作業実績報告(委任)」シート作成
				if _, err := f.NewSheet(OutsourceSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 勤務日行頭埋め
				for row := 20; row <= 50; row++ {
					if err := f.SetCellValue(OutsourceSheetName, "B"+strconv.Itoa(row), row-19); err != nil {
						assert.FailNow(t, err.Error())
					}
				}
				// 「御見積書 (委任)」シートを作らない
				if err := f.SaveAs(templatePath); err != nil {
					assert.FailNow(t, err.Error())
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEachExcelScheduleMaker(tt.fs, tt.order, tt.workdays, tt.fullName, tt.outputFilePath, tt.templatePath)
			if tt.prepareTemplate != nil {
				tt.prepareTemplate(tt.fs, tt.templatePath)
			}
			gotErr := e.MakeSchedule()

			assertErrorOrNoError(t, tt.wantErr, gotErr)
		})
	}
}

func TestEachExcelScheduleMaker_fillBeforeUseSheet(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		order          *Order
		prepareOutFile func() *excelize.File
		wantErr        bool
		assert         func(*excelize.File)
	}{
		{
			name: "正常系",
			order: &Order{
				ProjectCode:    "X12Y3456",
				DepartmentCode: "Z12345",
				Customer:       "A社",
				Issue:          "２０２４\u3000ＯＯＯ業務",
			},
			prepareOutFile: func() *excelize.File {
				outFile := excelize.NewFile()
				if _, err := outFile.NewSheet(BeforeUseSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				return outFile
			},
			wantErr: false,
			assert: func(outFile *excelize.File) {
				assertCellEq := func(cell string, expectedCellValue string) {
					gotCellValue, err := outFile.GetCellValue(BeforeUseSheetName, cell)
					if err != nil {
						assert.FailNow(t, err.Error())
					}
					assert.Equal(t, expectedCellValue, gotCellValue)
				}
				assertStridedStrEq := func(col int, row int, expectedStr string) {
					for i, expectedCh := range expectedStr {
						cell, err := excelize.CoordinatesToCellName(col+i, row)
						if err != nil {
							assert.FailNow(t, err.Error())
						}
						assertCellEq(cell, string(expectedCh))
					}
				}
				// PrjCD
				assertStridedStrEq(4, 18, "X12Y3456")
				// 担当部署CD
				assertStridedStrEq(4, 19, "Z12345")
				// 顧客名
				assertCellEq("D20", "A社")
				// 案件名
				assertCellEq("D21", "２０２４\u3000ＯＯＯ業務")
			},
		},
		{
			name: "該当シート無し",
			order: &Order{
				ProjectCode:    "X12Y3456",
				DepartmentCode: "Z12345",
				Customer:       "A社",
				Issue:          "２０２４\u3000ＯＯＯ業務",
			},
			prepareOutFile: func() *excelize.File {
				return excelize.NewFile()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEachExcelScheduleMaker(afero.NewMemMapFs(),
				tt.order,
				[]int{},
				"fullname",
				"outFilePath",
				"templateFilePath",
			)
			outFile := tt.prepareOutFile()
			defer outFile.Close()

			gotErr := e.fillBeforeUseSheet(outFile)

			assertErrorOrNoError(t, tt.wantErr, gotErr)
			if tt.assert != nil {
				tt.assert(outFile)
			}
		})
	}
}

func TestEachExcelScheduleMaker_fillOutsourceSheet(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		order          *Order
		workdays       []int
		fullName       string
		prepareOutFile func() *excelize.File
		wantErr        bool
		assert         func(outFile *excelize.File, workdays []int)
	}{
		{
			name: "正常系",
			order: &Order{
				OrderNo:        "123456",
				Issue:          "２０２４　ＯＯＯ業務",
				Representative: "I. K.",
				StartDate:      OrderDate{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				EndDate:        OrderDate{Time: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)},
			},
			workdays: []int{
				3, 4, 5,
				8, 9, 10, 11, 12,
				15, 16, 17, 18, 19,
				22, 23, 24, 25, 26,
				29, 30, 31,
			},
			fullName: "山田太郎",
			prepareOutFile: func() *excelize.File {
				outFile := excelize.NewFile()
				// シート作成
				if _, err := outFile.NewSheet(OutsourceSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 勤務日行頭埋め
				for row := 20; row <= 50; row++ {
					if err := outFile.SetCellValue(OutsourceSheetName, "B"+strconv.Itoa(row), row-19); err != nil {
						assert.FailNow(t, err.Error())
					}
				}
				return outFile
			},
			wantErr: false,
			assert: func(outFile *excelize.File, workdays []int) {
				assertCellEq := func(cell string, expectedCellValue string) {
					gotCellValue, err := outFile.GetCellValue(OutsourceSheetName, cell)
					if err != nil {
						assert.FailNow(t, err.Error())
					}
					assert.Equal(t, expectedCellValue, gotCellValue)
				}
				// 勤務表上部の年月
				assertCellEq("G9", "2024")
				assertCellEq("K9", "1")
				// 勤務表上部の発注番号、氏名
				assertCellEq("E14", "S-123456")
				assertCellEq("S16", "山田太郎")
				// 勤務表下部の担当責任者
				assertCellEq("S60", "I. K.")
				// 各勤務日
				for row := 20; row <= 50; row++ {
					if slices.Contains(workdays, row-19) {
						assertCellEq("C"+strconv.Itoa(row), "08:00")
						assertCellEq("N"+strconv.Itoa(row), "２０２４　ＯＯＯ業務")
					} else {
						assertCellEq("C"+strconv.Itoa(row), "")
						assertCellEq("N"+strconv.Itoa(row), "")
					}
				}
			},
		},
		{
			name: "該当シート無し",
			order: &Order{
				OrderNo:        "123456",
				Issue:          "２０２４　ＯＯＯ業務",
				Representative: "I. K.",
				StartDate:      OrderDate{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				EndDate:        OrderDate{Time: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)},
			},
			prepareOutFile: func() *excelize.File {
				return excelize.NewFile()
			},
			wantErr: true,
		},
		{
			name:  "日付であるはずのセルに数値以外が入っている",
			order: &Order{},
			prepareOutFile: func() *excelize.File {
				outFile := excelize.NewFile()
				// シート作成
				if _, err := outFile.NewSheet(OutsourceSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				// 勤務日行頭埋め(文字列)
				if err := outFile.SetCellValue(OutsourceSheetName, "B"+strconv.Itoa(20), "not a number"); err != nil {
					assert.FailNow(t, err.Error())
				}
				return outFile
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEachExcelScheduleMaker(afero.NewMemMapFs(),
				tt.order,
				tt.workdays,
				"山田太郎",
				"outFilePath",
				"templateFilePath",
			)
			outFile := tt.prepareOutFile()
			defer outFile.Close()

			gotErr := e.fillOutSourceSheet(outFile)
			assertErrorOrNoError(t, tt.wantErr, gotErr)

			if tt.assert != nil {
				tt.assert(outFile, tt.workdays)
			}
		})
	}
}

func TestEachExcelScheduleMaker_fillQuotDelegationSheet(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		order          *Order
		prepareOutFile func() *excelize.File
		wantErr        bool
		assert         func(*excelize.File)
	}{
		{
			name:  "正常系",
			order: &Order{Department: "開発二課"},
			prepareOutFile: func() *excelize.File {
				outFile := excelize.NewFile()
				if _, err := outFile.NewSheet(QuotationDelegationSheetName); err != nil {
					assert.FailNow(t, err.Error())
				}
				return outFile
			},
			wantErr: false,
			assert: func(outFile *excelize.File) {
				val, err := outFile.GetCellValue(QuotationDelegationSheetName, "I26")
				assert.NoError(t, err)
				assert.Equal(t, "開発二課", val)
			},
		},
		{
			name:  "該当シート無し",
			order: &Order{Department: "開発二課"},
			prepareOutFile: func() *excelize.File {
				return excelize.NewFile()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEachExcelScheduleMaker(
				afero.NewMemMapFs(),
				tt.order, []int{},
				"fullname",
				"outFilePath",
				"templateFilePath",
			)
			outFile := tt.prepareOutFile()
			defer outFile.Close()

			gotErr := e.fillQuotDelegationSheet(outFile)
			assertErrorOrNoError(t, tt.wantErr, gotErr)
			if tt.assert != nil {
				tt.assert(outFile)
			}
		})
	}
}

type createErrorFs struct {
	afero.Fs
}

func (c *createErrorFs) Create(name string) (afero.File, error) {
	return nil, errors.New("create error (for test)")
}
func TestEachExcelScheduleMaker_copyFile(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		prepareFs func() afero.Fs
		src       string
		dst       string
		wantErr   bool
	}{
		{
			name: "正常系",
			prepareFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				if _, err := fs.Create("src.json"); err != nil {
					assert.FailNow(t, err.Error())
				}
				return fs
			},
			src:     "src.json",
			dst:     "dst.json",
			wantErr: false,
		},
		{
			name:      "ファイルオープンエラー",
			prepareFs: afero.NewMemMapFs,
			src:       "non_existent_file",
			wantErr:   true,
		},
		{
			name: "コピー先フォルダ作成エラー",
			prepareFs: func() afero.Fs {
				return afero.NewReadOnlyFs(afero.NewMemMapFs())
			},
			dst:     "temp/dst.json",
			wantErr: true,
		},
		{
			name: "コピー先ファイル作成エラー",
			prepareFs: func() afero.Fs {
				return &createErrorFs{Fs: afero.NewMemMapFs()}
			},
			dst:     "temp/dst.json",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFs := tt.prepareFs()
			e := NewEachExcelScheduleMaker(testFs, nil, nil, "fullname",
				"outputFilePath", "templatePath")
			gotErr := e.copyFile(tt.src, tt.dst)

			assertErrorOrNoError(t, tt.wantErr, gotErr)
		})
	}
}
