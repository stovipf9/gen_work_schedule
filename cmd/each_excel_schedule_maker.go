package cmd

import (
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/xuri/excelize/v2"
)

type EachExcelScheduleMaker struct {
	fs               afero.Fs
	order            *Order
	workdays         []int
	fullName         string
	outputFilePath   string
	templateFilePath string
}

func NewEachExcelScheduleMaker(
	fs afero.Fs,
	order *Order,
	workdays []int,
	fullName string,
	outputFilePath string,
	templatePath string,
) *EachExcelScheduleMaker {
	return &EachExcelScheduleMaker{
		fs:               fs,
		order:            order,
		workdays:         workdays,
		fullName:         fullName,
		outputFilePath:   outputFilePath,
		templateFilePath: templatePath,
	}
}

func (e *EachExcelScheduleMaker) MakeSchedule() error {
	// テンプレートからコピー
	if err := e.copyFile(e.templateFilePath, e.outputFilePath); err != nil {
		return err
	}

	// 出力ファイルを開く
	outFile, err := excelize.OpenFile(e.outputFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 「ご使用の前に」シート
	if err := e.fillBeforeUseSheet(outFile); err != nil {
		return err
	}

	// 「システム要員作業実績報告(委任)」シート
	if err := e.fillOutSourceSheet(outFile); err != nil {
		return err
	}

	// 「御見積書 (委任)」シート
	if err := e.fillQuotDelegationSheet(outFile); err != nil {
		return err
	}

	// セル再計算を行い
	if err := outFile.UpdateLinkedValue(); err != nil {
		return err
	}
	// 編集結果を保存
	if err := outFile.Save(); err != nil {
		return err
	}

	return nil
}

func (e *EachExcelScheduleMaker) fillBeforeUseSheet(
	outFile *excelize.File,
) error {
	// 1セル1文字ずつ入力する情報
	for _, v := range []struct {
		leftmostCell string
		str          string
	}{
		{leftmostCell: "D18", str: e.order.ProjectCode},
		{leftmostCell: "D19", str: e.order.DepartmentCode},
	} {
		// 左端のセル名を取得
		col, row, err := excelize.CellNameToCoordinates(v.leftmostCell)
		if err != nil {
			return err
		}
		// 左から1文字ずつ記入
		for colShift, ch := range v.str {
			cellName, err := excelize.CoordinatesToCellName(
				col+colShift, row,
			)
			if err != nil {
				return err
			}
			if err := outFile.SetCellValue(
				BeforeUseSheetName, cellName, string(ch),
			); err != nil {
				return err
			}
		}
	}

	// 1セルに文字列として入力する情報
	for _, v := range []struct {
		cell string
		str  string
	}{
		{cell: "D20", str: e.order.Customer},
		{cell: "D21", str: e.order.Issue},
	} {
		if err := outFile.SetCellValue(
			BeforeUseSheetName, v.cell, v.str,
		); err != nil {
			return err
		}
	}

	return nil
}

func (e *EachExcelScheduleMaker) fillOutSourceSheet(
	outFile *excelize.File,
) error {

	// 勤務表の上部・下部
	for _, v := range []struct {
		cell  string
		value any
	}{
		{cell: "G9", value: e.order.StartDate.Year()},
		{cell: "K9", value: int(e.order.StartDate.Month())},
		{cell: "E14", value: "S-" + e.order.OrderNo},
		{cell: "S16", value: e.fullName},
		{cell: "B60", value: e.order.Department},
		{cell: "S60", value: e.order.Representative},
	} {
		if err := outFile.SetCellValue(OutsourceSheetName, v.cell, v.value); err != nil {
			return err
		}
	}

	// 各勤務日
	for row := 20; row <= 50; row++ {
		dayStr, err := outFile.GetCellValue(
			OutsourceSheetName, "B"+strconv.Itoa(row))
		if err != nil {
			return err
		}
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			return err
		}

		// 1行分埋める
		for _, v := range []struct {
			colName string
			value   any
		}{
			{colName: "C", value: time.Duration(8) * time.Hour}, //定時想定
			{colName: "N", value: e.order.Issue},
		} {
			// 勤務日でないならばセルを空とする
			if !slices.Contains(e.workdays, day) {
				v.value = nil
			}

			if err := outFile.SetCellValue(
				OutsourceSheetName, v.colName+strconv.Itoa(row), v.value,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *EachExcelScheduleMaker) fillQuotDelegationSheet(
	outFile *excelize.File,
) error {
	return outFile.SetCellValue(
		QuotationDelegationSheetName, "I26", e.order.Department,
	)
}

func (e *EachExcelScheduleMaker) copyFile(src string, dst string) error {
	sourceFile, err := e.fs.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// コピー先フォルダが存在しない場合は作成
	if err := e.fs.MkdirAll(filepath.Dir(dst), 0666); err != nil {
		return err
	}

	destFile, err := e.fs.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
