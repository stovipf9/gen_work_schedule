package main

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

func FirstSheet(
	outFile *excelize.File,
	record map[string]interface{},
) error {
	for _, colName := range []string{ProjectCodeColName, DepartmentCodeColName, CustomerColName, IssueColName} {
		if record[colName] == nil {
			return fmt.Errorf("value of %v is nil", colName)
		}
	}

	// 1セル1文字ずつ入力する情報
	for _, v := range []struct {
		leftmostCell string
		header       string
	}{
		{leftmostCell: "D18", header: ProjectCodeColName},
		{leftmostCell: "D19", header: DepartmentCodeColName},
	} {
		// 左端のセル名を取得
		col, row, err := excelize.CellNameToCoordinates(v.leftmostCell)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// 左から1文字ずつ記入
		for colShift, ch := range record[v.header].(string) {
			cellName, err := excelize.CoordinatesToCellName(
				col+colShift, row)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if err := outFile.SetCellValue(
				BeforeUseSheetName, cellName, string(ch)); err != nil {
				fmt.Println(err)
			}
		}
	}

	// 1セルに文字列として入力する情報
	for _, v := range []struct {
		cell   string
		header string
	}{
		{cell: "D20", header: CustomerColName},
		{cell: "D21", header: IssueColName},
	} {
		if err := outFile.SetCellValue(
			BeforeUseSheetName, v.cell, record[v.header].(string),
		); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func SecondSheet(
	outFile *excelize.File,
	record map[string]interface{},
	startDate time.Time,
	workDays []int,
) error {
	for _, colName := range []string{OrderNoColName, FullNameColName, IssueColName} {
		if record[colName] == nil {
			return fmt.Errorf("value %v is nil", colName)
		}
	}
	// 勤務表の上部
	for _, v := range []struct {
		cell  string
		value interface{}
	}{
		{cell: "G9", value: startDate.Year()},
		{cell: "K9", value: int(startDate.Month())},
		{cell: "E14", value: "S-" + record[OrderNoColName].(string)},
		{cell: "S16", value: record[FullNameColName].(string)},
	} {
		if err := outFile.SetCellValue(OutsourceSheetName, v.cell, v.value); err != nil {
			fmt.Println(err)
		}
	}

	// 各勤務日
	for row := 20; row <= 50; row++ {
		dayStr, err := outFile.GetCellValue(
			OutsourceSheetName, "B"+strconv.Itoa(row))
		if err != nil {
			fmt.Println(err)
			continue
		}
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 1行分埋める
		for _, v := range []struct {
			colName string
			value   interface{}
		}{
			{colName: "C", value: time.Duration(8) * time.Hour},
			{colName: "N", value: record[IssueColName].(string)},
		} {
			// 勤務日でないならばセルを空とする
			if !slices.Contains(workDays, day) {
				v.value = nil
			}

			if err := outFile.SetCellValue(
				OutsourceSheetName, v.colName+strconv.Itoa(row), v.value,
			); err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}

func createEachSpreadSheet(
	record map[string]interface{},
	startDate time.Time,
	workDays []int,
	pathOf PathOf,
) error {
	outPath := pathOf.OutputDir + fmt.Sprintf(
		"/S-%s【%s】%s_%04d%02d.xlsx",
		record[OrderNoColName],
		record[CompanyColName],
		record[FamilyNameColName],
		startDate.Year(), startDate.Month())
	if err := copyFile(pathOf.TemplateXlsx, outPath); err != nil {
		return err
	}
	outFile, err := excelize.OpenFile(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 「ご使用の前に」シート
	if err := FirstSheet(outFile, record); err != nil {
		return err
	}

	// 「システム要員作業実績報告(委任)」シート
	if err := SecondSheet(outFile, record, startDate, workDays); err != nil {
		return err
	}

	// セル再計算の後、編集結果を保存
	if err := outFile.UpdateLinkedValue(); err != nil {
		return err
	}
	if err := outFile.Save(); err != nil {
		return err
	}

	fmt.Printf("%s is saved.\n", outPath)
	return nil
}
