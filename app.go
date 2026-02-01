package main

import (
	"context"
	"gen_work_schedule/cmd"
	"os/exec"
	"reflect"
	"runtime"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type TableColumn struct {
	Field string `json:"field"`
	Label string `json:"label"`
}

func getColumnNames[T any]() []TableColumn {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	columns := []TableColumn{}
	for i := range typ.NumField() {
		field := typ.Field(i)
		columns = append(columns, TableColumn{
			Field: field.Tag.Get("json"),
			Label: field.Tag.Get("csv"),
		})
	}
	return columns
}

// WailsApp struct
type WailsApp struct {
	ctx    context.Context
	logger *cmd.SLogger

	OrderService         *cmd.OrderService
	RosterService        *cmd.RosterService
	FileSelectionService *WailsFileSelectionService
	GenerateService      *cmd.GenerateService
	wailsCacheService    *WailsCacheService

	paths map[cmd.PathKey]string
}

func (a *WailsApp) SetPaths(paths map[cmd.PathKey]string) {
	a.paths = paths
}

func (WailsApp) GetOrderColumnNames() []TableColumn {
	return getColumnNames[cmd.Order]()
}

func (WailsApp) GetRosterColumnNames() []TableColumn {
	return getColumnNames[cmd.RosterRecord]()
}

func (a *WailsApp) LoadOrders(path string) ([]cmd.Order, error) {
	ret, err := a.OrderService.LoadOrders(path)
	return ret, a.logger.Log(cmd.ERROR, err)
}

func (a *WailsApp) LoadRosterRecords(path string) (map[string]string, error) {
	ret, err := a.RosterService.LoadRosterRecords(path)
	return ret, a.logger.Log(cmd.ERROR, err)
}

func (a *WailsApp) Generate(paths map[cmd.PathKey]string) (
	eachResults map[string]error,
	criticalError error,
) {
	orders, err := a.OrderService.LoadOrders(paths[cmd.OrderXlsx])
	if err != nil {
		return eachResults, a.logger.Log(cmd.ERROR, err)
	}
	rosterRecords, err := a.RosterService.LoadRosterRecords(paths[cmd.NamesCSV])
	if err != nil {
		return eachResults, a.logger.Log(cmd.ERROR, err)
	}

	ret, err := a.GenerateService.Generate(orders, rosterRecords,
		paths[cmd.OutputDir], paths[cmd.TemplateXlsx])
	return ret, a.logger.Log(cmd.WARNING, err)
}

func (a *WailsApp) SelectDirectory() (string, error) {
	ret, err := a.FileSelectionService.SelectDirectory()
	return ret, a.logger.Log(cmd.ERROR, err)
}

func (a *WailsApp) SelectFile(extension string) (string, error) {
	ret, err := a.FileSelectionService.SelectFile(extension)
	return ret, a.logger.Log(cmd.ERROR, err)
}

func (a *WailsApp) LoadCachedPaths() (map[cmd.PathKey]string, error) {
	paths, err := a.wailsCacheService.LoadCachedPaths()
	a.SetPaths(paths)
	return paths, a.logger.Log(cmd.INFO, err)
}

func WithLogger(logger *cmd.SLogger) func(*WailsApp) {
	return func(a *WailsApp) {
		a.logger = logger
	}
}

func WithOrderService(orderService *cmd.OrderService) func(*WailsApp) {
	return func(a *WailsApp) {
		a.OrderService = orderService
	}
}

func WithRosterService(rosterService *cmd.RosterService) func(*WailsApp) {
	return func(a *WailsApp) {
		a.RosterService = rosterService
	}
}

func WithFileSelectionService(fileSelectionService *WailsFileSelectionService) func(*WailsApp) {
	return func(a *WailsApp) {
		fileSelectionService.setContext(&a.ctx)
		a.FileSelectionService = fileSelectionService
	}
}

func WithGenerateService(generateService *cmd.GenerateService) func(*WailsApp) {
	return func(a *WailsApp) {
		a.GenerateService = generateService
	}
}

func WithWailsCacheService(wailsCacheService *WailsCacheService) func(*WailsApp) {
	return func(a *WailsApp) {
		a.wailsCacheService = wailsCacheService
	}
}

func NewWailsApp(serviceSetters ...func(*WailsApp)) *WailsApp {
	app := &WailsApp{paths: make(map[cmd.PathKey]string)}
	for _, serviceSetter := range serviceSetters {
		serviceSetter(app)
	}
	return app
}

func (a *WailsApp) OpenFile(path string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		command = exec.Command("open", path)
	default: // "linux", "freebsd", "openbsd", "netbsd"
		command = exec.Command("xdg-open", path)
	}

	return a.logger.Log(cmd.ERROR, command.Start())
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *WailsApp) startup(ctx context.Context) {
	a.ctx = ctx

	// 画面位置サイズをキャッシュからロード
	if windowPosSize, err := a.wailsCacheService.LoadCachedWindowPosSize(); err == nil {
		wailsruntime.WindowSetPosition(a.ctx,
			windowPosSize.Pos.X, windowPosSize.Pos.Y)
		wailsruntime.WindowSetSize(a.ctx, windowPosSize.Size.W, windowPosSize.Size.H)
	} else {
		a.logger.Log(cmd.WARNING, err)
	}
}

func (a *WailsApp) beforeClose(ctx context.Context) (prevent bool) {
	// 画面位置サイズをキャッシュファイルに保存
	x, y := wailsruntime.WindowGetPosition(a.ctx)
	w, h := wailsruntime.WindowGetSize(a.ctx)
	a.logger.Log(cmd.WARNING, a.wailsCacheService.CacheWindowPosSize(
		WindowPosSize{
			Pos{x, y},
			Size{w, h},
		},
	))

	return false
}

func (a *WailsApp) shutdown(_ context.Context) {
	// パス情報をキャッシュ
	_ = a.logger.Log(cmd.WARNING, a.wailsCacheService.CachePaths(a.paths))

}
