package main

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type OrderInfo struct {
	Company        string
	FamilyName     string
	OrderNo        string
	Issue          string
	ProjectCode    string
	Customer       string
	DepartmentCode string
	// Amount         uint
	// Unit           string
	StartDate      time.Time
	EndDate        time.Time
	FullName       string
	Representative string
}

func ToOrderInfo(record map[string]interface{}) (OrderInfo, error) {
	var ret OrderInfo

	// 文字列型
	for colName, retv := range map[string]*string{
		CompanyColName:        &ret.Company,
		FamilyNameColName:     &ret.FamilyName,
		OrderNoColName:        &ret.OrderNo,
		IssueColName:          &ret.Issue,
		ProjectCodeColName:    &ret.ProjectCode,
		CustomerColName:       &ret.Customer,
		DepartmentCodeColName: &ret.DepartmentCode,
		FullNameColName:       &ret.FullName,
		RepresentativeColName: &ret.Representative,
	} {
		if v, ok := record[colName]; ok {
			if strVal, ok := v.(string); ok {
				*retv = strVal
			} else {
				return OrderInfo{}, fmt.Errorf("%v is not string", strVal)
			}
		} else {
			return OrderInfo{}, fmt.Errorf("there is no key: %v", colName)
		}
	}

	// 日時型
	for colName, retv := range map[string]*time.Time{
		StartDateColName: &ret.StartDate,
		EndDateColName:   &ret.EndDate,
	} {
		if v, ok := record[colName]; ok {
			if strVal, ok := v.(string); ok {
				dt, err := time.Parse("20060102", strVal)
				if err != nil {
					return OrderInfo{}, err
				}
				*retv = dt
			} else {
				return OrderInfo{}, fmt.Errorf("%v is not string", v)
			}
		} else {
			return OrderInfo{}, fmt.Errorf("there is no key: %v", colName)
		}
	}

	return ret, nil
}

func FirstSheet(
	outFile *excelize.File,
	orderInfo OrderInfo,
) error {
	// 1セル1文字ずつ入力する情報
	for _, v := range []struct {
		leftmostCell string
		str          string
	}{
		{leftmostCell: "D18", str: orderInfo.ProjectCode},
		{leftmostCell: "D19", str: orderInfo.DepartmentCode},
	} {
		// 左端のセル名を取得
		col, row, err := excelize.CellNameToCoordinates(v.leftmostCell)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// 左から1文字ずつ記入
		for colShift, ch := range v.str {
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
		cell string
		str  string
	}{
		{cell: "D20", str: orderInfo.Customer},
		{cell: "D21", str: orderInfo.Issue},
	} {
		if err := outFile.SetCellValue(
			BeforeUseSheetName, v.cell, v.str,
		); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func SecondSheet(
	outFile *excelize.File,
	orderInfo OrderInfo,
	cacheDirIF CacheDirInterface,
) error {
	workDays, err := retrieveWorkDays(orderInfo.StartDate, orderInfo.EndDate, cacheDirIF)
	if err != nil {
		fmt.Println(err)
	}

	// 勤務表の上部・下部
	for _, v := range []struct {
		cell  string
		value interface{}
	}{
		{cell: "G9", value: orderInfo.StartDate.Year()},
		{cell: "K9", value: int(orderInfo.StartDate.Month())},
		{cell: "E14", value: "S-" + orderInfo.OrderNo},
		{cell: "S16", value: orderInfo.FullName},
		{cell: "S60", value: orderInfo.Representative},
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
			{colName: "N", value: orderInfo.Issue},
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
	pathOf PathOf,
	cacheDirIF CacheDirInterface,
) error {
	orderInfo, err := ToOrderInfo(record)
	if err != nil {
		return err
	}

	outPath := pathOf.OutputDir + fmt.Sprintf(
		"/S-%s【%s】%s_%04d%02d.xlsx",
		orderInfo.OrderNo,
		orderInfo.Company,
		orderInfo.FamilyName,
		orderInfo.StartDate.Year(), orderInfo.StartDate.Month())
	if err := copyFile(pathOf.TemplateXlsx, outPath); err != nil {
		return err
	}
	outFile, err := excelize.OpenFile(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 「ご使用の前に」シート
	if err := FirstSheet(outFile, orderInfo); err != nil {
		return err
	}

	// 「システム要員作業実績報告(委任)」シート
	if err := SecondSheet(outFile, orderInfo, cacheDirIF); err != nil {
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
