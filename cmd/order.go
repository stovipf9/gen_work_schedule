package cmd

import (
	"fmt"
	"time"
)

// RFC2339形式の日時文字列ではないため、
// シリアライズのためにカスタム構造体を定義
type OrderDate struct {
	time.Time
}

func (od *OrderDate) UnmarshalCSV(data []byte) error {
	tt, err := time.Parse("20060102", string(data))
	if err != nil {
		return err
	}
	od.Time = tt
	return nil
}

type Order struct {
	Company        string    `json:"company" csv:"会社"`
	FamilyName     string    `json:"familyName" csv:"姓"`
	OrderNo        string    `json:"orderNo" csv:"発注番号"`
	Issue          string    `json:"issue" csv:"案件名"`
	ProjectCode    string    `json:"projectCode" csv:"PrjCD"`
	Customer       string    `json:"customer" csv:"顧客名"`
	DepartmentCode string    `json:"departmentCode" csv:"担当部署CD"`
	Amount         uint      `json:"amount" csv:"数量"`
	Unit           string    `json:"unit" csv:"単位"`
	StartDate      OrderDate `json:"startDate" csv:"作業開始日"`
	EndDate        OrderDate `json:"endDate" csv:"作業終了日"`
	Department     string    `json:"department" csv:"納入物の提出場所"`
	Representative string    `json:"representative" csv:"担当責任者"`
	SequenceNo     uint
}

func (o Order) GetOutFileName() string {
	return fmt.Sprintf(
		"S-%s_%04d%02d_%d_【%s】%s.xlsx",
		o.OrderNo,
		o.StartDate.Year(), o.StartDate.Month(),
		o.SequenceNo,
		o.Company, o.FamilyName,
	)
}
