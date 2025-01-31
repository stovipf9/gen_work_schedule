package main

const (
	HolidaysURL           = "https://holidays-jp.shogo82148.com"
	BeforeUseSheetName    = "ご使用の前に"
	OutsourceSheetName    = "システム要員作業実績報告(委任)"
	CompanyColName        = "会社"
	FamilyNameColName     = "姓"
	OrderNoColName        = "発注番号"
	ProjectCodeColName    = "PrjCD"
	DepartmentCodeColName = "担当部署CD"
	CustomerColName       = "顧客名"
	IssueColName          = "案件名"
	StartDateColName      = "作業開始日"
	EndDateColName        = "作業終了日"
	FullNameColName       = "姓名"
	RepresentativeColName = "担当責任者"
)

func CacheDir(cacheDirIF CacheDirInterface) (string, error) {
	path, err := cacheDirIF.UserCacheDir()
	if err != nil {
		return "", err
	}
	return path + "/GenWorkSchedule/", nil
}

func HolidaysJSONPath(cacheDirIF CacheDirInterface) (string, error) {
	cacheDir, err := CacheDir(cacheDirIF)
	if err != nil {
		return "", err
	}
	return cacheDir + "holidays.json", nil
}

func PathsJSONPath(cacheDirIF CacheDirInterface) (string, error) {
	cacheDir, err := CacheDir(cacheDirIF)
	if err != nil {
		return "", err
	}
	return cacheDir + "paths.json", nil
}
