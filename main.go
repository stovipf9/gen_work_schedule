package main

import (
	"embed"
	"gen_work_schedule/cmd"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	osFs := afero.NewOsFs()
	// TODO: エラーの握りつぶしを無くす
	execPath, _ := os.Executable()
	sLogger := cmd.NewSLogger(afero.NewBasePathFs(osFs, filepath.Dir(execPath)), "app.log")
	defer sLogger.Close()

	userCacheDir, err := os.UserCacheDir()
	sLogger.Log(cmd.WARNING, err)

	wailsCacheService := NewWailsCacheService(
		cmd.NewCacheService(afero.NewBasePathFs(osFs, userCacheDir), "gen_work_schedule"),
	)
	app := NewWailsApp(
		WithOrderService(cmd.NewOrderService()),
		WithRosterService(cmd.NewRosterService(osFs)),
		WithFileSelectionService(NewWailsFileSelectionService()),
		WithGenerateService(cmd.NewGenerateService(
			cmd.NewRetrieveWorkdaysService(
				cmd.NewHttpGetter(), wailsCacheService, sLogger),
			cmd.NewScheduleMakerFactoryWithFs(osFs),
		)),
		WithWailsCacheService(wailsCacheService),
		WithLogger(sLogger),
	)

	defer func() {
		if r := recover(); r != nil {
			sLogger.Info("Panic recovered in main:")
		}
	}()

	// Create application with options
	sLogger.Log(cmd.ERROR, wails.Run(&options.App{
		Title:  "gen_work_schedule",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		EnableDefaultContextMenu: false,
		OnStartup:                app.startup,
		OnBeforeClose:            app.beforeClose,
		OnShutdown:               app.shutdown,
		Bind: []any{
			app,
		},
		Logger:   sLogger,
		LogLevel: logger.TRACE,
		EnumBind: []any{
			[]struct {
				Value  cmd.PathKey
				TSName string
			}{
				{cmd.OrderXlsx, "ORDER_XLSX"},
				{cmd.NamesCSV, "NAMES_CSV"},
				{cmd.TemplateXlsx, "TEMPLATE_XLSX"},
				{cmd.OutputDir, "OUTPUT_DIR"},
			},
		},
	}))

}
