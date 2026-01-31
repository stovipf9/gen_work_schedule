package main

import (
	"gen_work_schedule/cmd"
)

const (
	WINDOW_POS_SIZE_JSON = "window_pos_size.json"
)

// CacheServiceに対してウィンドウサイズをキャッシュするメソッドを追加する
type WailsCacheService struct {
	*cmd.CacheService
}

func NewWailsCacheService(cacheService *cmd.CacheService) *WailsCacheService {
	return &WailsCacheService{CacheService: cacheService}
}

func (wc WailsCacheService) LoadCachedWindowPosSize() (WindowPosSize, error) {
	return cmd.LoadCachedData[WindowPosSize](
		wc.CacheDirFs(),
		WINDOW_POS_SIZE_JSON,
	)
}

func (wc WailsCacheService) CacheWindowPosSize(windowRect WindowPosSize) error {
	return cmd.CacheData(
		wc.CacheDirFs(),
		windowRect,
		WINDOW_POS_SIZE_JSON,
	)
}
