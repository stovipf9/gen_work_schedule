package main

import (
	"io"
	"os"

	"github.com/ncruces/zenity"
)

type AppInterfaces struct {
	ZenityIF    ZenityInterface
	CacheDirIF  CacheDirInterface
	InputReader io.Reader
}

type ZenityInterface interface {
	SelectFile(extension string) (string, error)
	SelectDirectory() (string, error)
}

type RealZenity struct{}

func (RealZenity) SelectFile(extension string) (string, error) {
	return zenity.SelectFile(
		zenity.FileFilters{
			// 指定拡張子
			zenity.FileFilter{
				Patterns: []string{"*." + extension},
			},
			// 全てのファイル
			zenity.FileFilter{
				Patterns: []string{"*"},
			},
		},
		// カレントディレクトリから始める
		zenity.Filename("."),
	)
}

func (RealZenity) SelectDirectory() (string, error) {
	return zenity.SelectFile(
		zenity.Directory(),
		// カレントディレクトリから始める
		zenity.Filename("."),
	)
}

type CacheDirInterface interface {
	UserCacheDir() (string, error)
}

type RealCacheDirIF struct{}

func (RealCacheDirIF) UserCacheDir() (string, error) {
	return os.UserCacheDir()
}
