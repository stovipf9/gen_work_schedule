package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
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
			errors <- createEachSpreadSheet(r, pathOf, appIF.CacheDirIF)
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

func init() {
	if os.Getenv("DISABLE_EXTERNAL_HTTP") == "true" {
		http.DefaultTransport = &http.Transport{
			DisableKeepAlives: true,
			Dial: func(_, _ string) (net.Conn, error) {
				return nil, fmt.Errorf("外部HTTP接続は禁止されています")
			},
		}
	}
}
