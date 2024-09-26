package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

func joinFullNames(namesPath string, orderDf dataframe.DataFrame) (dataframe.DataFrame, error) {
	// 名簿を読み込み左結合
	csvFile, err := os.Open(namesPath)
	if err != nil {
		return orderDf, err
	}
	defer csvFile.Close()

	csvStr, err := io.ReadAll(csvFile)
	if err != nil {
		return orderDf, err
	}

	if !slices.Contains(orderDf.Names(), FamilyNameColName) {
		return orderDf, fmt.Errorf(`発注情報に"%s"の列がありません`, FamilyNameColName)
	}
	return orderDf.LeftJoin(
		dataframe.ReadCSV(
			strings.NewReader(string(csvStr))), FamilyNameColName), nil
}
