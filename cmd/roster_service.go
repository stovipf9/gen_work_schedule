package cmd

import (
	"fmt"

	"github.com/jszwec/csvutil"
	"github.com/samber/lo"
	"github.com/spf13/afero"
)

type RosterRecord struct {
	FamilyName string `json:"familyName" csv:"姓"`
	FullName   string `json:"fullName" csv:"姓名"`
}

type RosterService struct {
	fs afero.Fs
}

func NewRosterService(fs afero.Fs) *RosterService {
	return &RosterService{fs: afero.NewReadOnlyFs(fs)}
}

func (r RosterService) LoadRosterRecords(path string) (map[string]string, error) {
	ZERO := make(map[string]string)

	fileBytes, err := afero.ReadFile(r.fs, path)
	if err != nil {
		return ZERO, err
	}

	var records []RosterRecord
	if err := csvutil.Unmarshal(fileBytes, &records); err != nil {
		return ZERO, err
	}

	// 姓の列の一意性チェック
	if len(lo.FindDuplicatesBy(records, func(r RosterRecord) string { return r.FamilyName })) > 0 {
		return ZERO, fmt.Errorf("姓の列で重複する値が存在します。")
	}

	// マップ化
	retMap := make(map[string]string)
	for _, record := range records {
		retMap[record.FamilyName] = record.FullName
	}

	return retMap, nil
}
