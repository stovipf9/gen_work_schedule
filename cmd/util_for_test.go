package main

import (
	"errors"
	mock_main "gen_work_schedule/mocks"
	"io"
	"os"
	"testing"

	"github.com/go-gota/gota/dataframe"
	"go.uber.org/mock/gomock"
)

type MockAppInterfaces struct {
	mz *mock_main.MockZenityInterface
	mc *mock_main.MockCacheDirInterface
	mi io.Reader
}

type errorResponseReader struct{}

func (errorResponseReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

const (
	TestDirPath          = "../tests"
	TestUserCacheDirName = TestDirPath + "/cache"
)

func holidaysJSONPathGetterByMock(ctrl *gomock.Controller) string {
	m := mock_main.NewMockCacheDirInterface(ctrl)
	m.EXPECT().UserCacheDir().Return(TestUserCacheDirName, nil)
	holidaysJSONPath, _ := HolidaysJSONPath(m)

	return holidaysJSONPath
}

func prepareCacheJSONByMock(
	t *testing.T,
	jsonPathGetter func(
		ctrl *gomock.Controller,
	) string,
	json string,
) {
	mCtrl := gomock.NewController(t)
	defer mCtrl.Finish()

	f, _ := os.OpenFile(
		jsonPathGetter(mCtrl),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	defer f.Close()

	if _, err := f.WriteString(json); err != nil {
		t.Errorf("error in creating json: %v\n", err)
	}
}

func compareDataframes(df1, df2 dataframe.DataFrame) bool {
	// 次元をチェックする
	if df1.Nrow() != df2.Nrow() || df1.Ncol() != df2.Ncol() {
		return false
	}

	compareSlices := func(s1, s2 []string) bool {
		if len(s1) != len(s2) {
			return false
		}

		for i := range s1 {
			if s1[i] != s2[i] {
				return false
			}
		}
		return true
	}

	// 列名をチェックする
	if !compareSlices(df1.Names(), df2.Names()) {
		return false
	}

	// 値をチェックする
	for i := 0; i < df1.Nrow(); i++ {
		if !compareSlices(df1.Records()[i], df2.Records()[i]) {
			return false
		}
	}
	return true
}
