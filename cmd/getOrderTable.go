package main

import (
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/xuri/excelize/v2"
)

func getOrderTable(path string) (dataframe.DataFrame, error) {
	// スプレッドシートを開く
	f, err := excelize.OpenFile(path)
	if err != nil {
		return dataframe.DataFrame{}, err
	}
	defer f.Close()

	// 最初のシートのデータの範囲を取得する
	sheetName := f.GetSheetName(0)
	dim, err := f.GetSheetDimension(sheetName)
	if err != nil {
		return dataframe.DataFrame{}, err
	}
	topLeftCol, topLeftRow, err := excelize.CellNameToCoordinates(strings.Split(dim, ":")[0])
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	// 最初のシートのテーブルを行ごとに読み込む
	var records [][]string
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return dataframe.DataFrame{}, err
	}
	for _, row := range rows[topLeftRow-1:] {
		// 空の行は読み飛ばす
		// (GetSheetDimension()の不具合か、空行にもかかわらず範囲に含まれる場合があるため)
		if len(strings.Join(row, "")) == 0 {
			continue
		}

		var record []string
		for _, cell := range row[topLeftCol-1:] {
			// セルの値の両端の空白, CR, LFを除去した値を格納する
			record = append(record, strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(cell)))
		}

		records = append(records, record)
	}

	return dataframe.LoadRecords(
		records, dataframe.DetectTypes(false),
		dataframe.NaNValues([]string{})), nil
}
