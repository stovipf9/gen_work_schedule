package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathOf struct {
	// 発注情報
	OrderXlsx string
	// 名簿
	NamesCSV string
	// 勤務表テンプレート
	TemplateXlsx string
	// 出力ディレクトリ
	OutputDir string
}

func confirm(reader io.Reader, prompt string) bool {
	bufReader := bufio.NewReader(reader)
	for {
		fmt.Printf("%s [y/n]: ", prompt)
		response, err := bufReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func readPathsFromJSON(cacheDirIF CacheDirInterface) (PathOf, error) {
	pathsJSONPath, err := PathsJSONPath(cacheDirIF)
	if err != nil {
		return PathOf{}, err
	}
	// 前回実行時のパス情報を開く
	pathsCacheFile, err := os.Open(pathsJSONPath)
	if err != nil {
		return PathOf{}, err
	}
	defer pathsCacheFile.Close()

	// 前回実行時のパス情報を読み込む
	pathsCacheBytes, err := io.ReadAll(pathsCacheFile)
	if err != nil {
		return PathOf{}, err
	}

	var pathOf PathOf
	if err := json.NewDecoder(strings.NewReader(string(pathsCacheBytes))).Decode((&pathOf)); err != nil {
		return PathOf{}, err
	}

	// 入力情報となるパスについては、ファイル自身が存在しないものは捨てる
	for _, path := range []*string{
		&pathOf.OrderXlsx,
		&pathOf.NamesCSV,
		&pathOf.TemplateXlsx,
	} {
		if !fileExists(*path) {
			*path = ""
		}
	}

	return pathOf, nil
}

func inputPaths(appIF *AppInterfaces) (PathOf, error) {
	// 前回実行時のパス情報を読み込む
	pathOf, _ := readPathsFromJSON(appIF.CacheDirIF)

	// 発注情報の入力を促す
	bufReader := bufio.NewReader(appIF.InputReader)
	for _, v := range []struct {
		fileDescription string
		extension       string
		path            *string
	}{
		{
			fileDescription: "発注情報が記載されたxlsxファイル",
			extension:       "xlsx",
			path:            &pathOf.OrderXlsx,
		},
		{
			fileDescription: "名簿となるcsvファイル(姓,姓名のテーブル)",
			extension:       "csv",
			path:            &pathOf.NamesCSV,
		},
		{
			fileDescription: "勤務表テンプレートとなるxlsxファイル",
			extension:       "xlsx",
			path:            &pathOf.TemplateXlsx,
		},
		{
			fileDescription: "各作業の勤務表出力ディレクトリ",
			path:            &pathOf.OutputDir,
		},
	} {
		doUseCache := false
		if *v.path != "" {
			doUseCache = confirm(
				appIF.InputReader,
				fmt.Sprintf(
					"%sは\n %s\nでよろしいですか?",
					v.fileDescription, *v.path,
				),
			)
		}
		if !doUseCache {
			fmt.Printf("%sを選択してください。(Enter押下でファイル選択に進みます)", v.fileDescription)
			_, err := bufReader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}
			if v.extension != "" {
				*v.path, err = appIF.ZenityIF.SelectFile(v.extension)
			} else {
				*v.path, err = appIF.ZenityIF.SelectDirectory()
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(*v.path)
		}
	}

	// キャッシュにパス情報を記録する
	jsonPath, err := PathsJSONPath(appIF.CacheDirIF)
	if err != nil {
		fmt.Println(err)
	} else {
		jsonBytes, err := json.Marshal(pathOf)
		if err != nil {
			fmt.Println(err)
		} else if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
			fmt.Println(err)
		}
	}

	return pathOf, nil
}
