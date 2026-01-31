package main

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WailsFileSelectionService struct {
	ctx *context.Context
}

func NewWailsFileSelectionService() *WailsFileSelectionService {
	return &WailsFileSelectionService{}
}

func (w *WailsFileSelectionService) setContext(ctx *context.Context) {
	w.ctx = ctx
}

func (w WailsFileSelectionService) SelectFile(extension string) (string, error) {
	return runtime.OpenFileDialog(
		*w.ctx,
		runtime.OpenDialogOptions{
			DefaultDirectory: ".",
			Filters: []runtime.FileFilter{
				{ // 指定拡張子
					Pattern: "*." + extension,
				},
				{ // 全てのファイル
					Pattern: "*",
				},
			},
			CanCreateDirectories: true,
		},
	)
}

func (w WailsFileSelectionService) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(
		*w.ctx,
		runtime.OpenDialogOptions{
			DefaultDirectory:     ".",
			DefaultFilename:      "results_" + time.Now().Format("2006-01-02"),
			CanCreateDirectories: true,
		},
	)
}
