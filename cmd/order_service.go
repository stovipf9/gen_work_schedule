package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/jszwec/csvutil"
	"github.com/xuri/excelize/v2"
)

type OrderService struct{}

func NewOrderService() *OrderService {
	return &OrderService{}
}

func (OrderService) LoadOrders(path string) (
	orders []Order, _ error,
) {
	// スプレッドシートを開く
	f, err := excelize.OpenFile(path)
	if err != nil {
		return orders, err
	}
	defer f.Close()

	// 最初のシートのデータの範囲を取得する
	sheetName := f.GetSheetName(0)
	dim, err := f.GetSheetDimension(sheetName)
	if err != nil {
		return orders, err
	}
	topLeftCol, topLeftRow, err := excelize.CellNameToCoordinates(strings.Split(dim, ":")[0])
	if err != nil {
		return orders, err
	}

	// 最初のシートのテーブルを行ごとに読み込む
	// デシリアライズの簡単さのため、メモリ上CSV化してから処理する
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return orders, err
	}
	if len(rows) == 0 {
		return orders, fmt.Errorf("発注情報のテーブルサイズが0")
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	isWritten := false
	for _, row := range rows[topLeftRow-1:] {
		// 空の行は読み飛ばす
		// (GetSheetDimension()の不具合か、空行にもかかわらず範囲に含まれる場合があるため)
		if len(strings.Join(row, "")) == 0 {
			if !isWritten {
				// 前方の空行は読み飛ばすのみ
				continue
			} else {
				// 後方の空行は読み取りを終える
				break
			}
		}

		var trimmedRow []string
		for _, cell := range row[topLeftCol-1:] {
			// セルの値の両端の空白, CR, LFを除去した値を格納する
			trimmedRow = append(trimmedRow,
				strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(cell)))
		}
		if err := w.Write(trimmedRow); err != nil {
			return orders, err
		}
		isWritten = true
	}
	w.Flush()

	// CSVデータをデシリアライズ
	if err := csvutil.Unmarshal(buf.Bytes(), &orders); err != nil {
		return orders, err
	}

	return orders, nil
}
