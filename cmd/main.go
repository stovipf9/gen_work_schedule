package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

func run(appIF *AppInterfaces) error {
	// キャッシュディレクトリを作成する
	if err := makeCacheDir(appIF.CacheDirIF); err != nil {
		fmt.Println(err)
	}

	// 発注情報、名簿、テンプレート、出力フォルダをユーザに入力してもらう
	pathOf, err := inputPaths(appIF)
	if err != nil {
		return err
	}

	// 発注情報を取り込む
	orderDf, err := getOrderTable(pathOf.OrderXlsx)
	if err != nil {
		return err
	}

	// 名簿を発注情報に合併
	joinedDf, err := joinFullNames(pathOf.NamesCSV, orderDf)
	if err != nil {
		return err
	}

	// 発注情報より作業開始日を読み取る
	var startDate time.Time
	for _, startDateStr := range joinedDf.Col(StartDateColName).Records() {
		startDate, err = time.Parse("20060102", startDateStr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		break
	}
	if startDate.IsZero() {
		return fmt.Errorf("%sがすべて不正です", StartDateColName)
	}

	// 当月の勤務日を求める
	workDays, err := retrieveWorkDays(startDate, appIF.CacheDirIF)
	if err != nil {
		fmt.Println(err)
	}

	// 出力先ディレクトリが存在しない場合は作成する
	if err := os.MkdirAll(pathOf.OutputDir, 0666); err != nil {
		return err
	}

	// 発注情報についてループ
	errors := make(chan error, joinedDf.Nrow())
	var waitGroup sync.WaitGroup
	for _, record := range joinedDf.Maps() {
		waitGroup.Add(1)
		go func(r map[string]interface{}) {
			defer waitGroup.Done()
			errors <- createEachSpreadSheet(r, startDate, workDays, pathOf)
		}(record)
	}
	go func() {
		waitGroup.Wait()
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("完了しました。Enterで終了します。")
	_, err = bufio.NewReader(appIF.InputReader).ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func main() {
	if err := run(
		&AppInterfaces{
			ZenityIF:    RealZenity{},
			CacheDirIF:  RealCacheDirIF{},
			InputReader: os.Stdin,
		},
	); err != nil {
		fmt.Println(err)
	}
}
