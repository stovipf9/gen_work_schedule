package cmd

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/afero"
)

type WorkdaysRetriever interface {
	RetrieveWorkdays(
		startDate time.Time,
		endDate time.Time,
	) ([]int, error)
}

type EachScheduleMaker interface {
	MakeSchedule() error
}
type EachScheduleMakerFactory func(
	order *Order,
	workdays []int,
	fullName string,
	outputFilePath string,
	templatePath string,
) EachScheduleMaker

func NewScheduleMakerFactoryWithFs(fs afero.Fs) EachScheduleMakerFactory {
	return func(
		order *Order,
		workdays []int,
		fullName string,
		outputFilePath string,
		templatePath string,
	) EachScheduleMaker {
		return NewEachExcelScheduleMaker(
			fs,
			order, workdays, fullName, outputFilePath, templatePath,
		)
	}
}

type GenerateService struct {
	workdaysRetriever    WorkdaysRetriever
	newEachScheduleMaker EachScheduleMakerFactory
}

func NewGenerateService(
	workdaysRetriever WorkdaysRetriever,
	newEachScheduleMaker EachScheduleMakerFactory,
) *GenerateService {
	return &GenerateService{
		workdaysRetriever:    workdaysRetriever,
		newEachScheduleMaker: newEachScheduleMaker,
	}
}

func (g *GenerateService) Generate(
	orders []Order,
	rosterRecords map[string]string,
	outputDirPath string,
	templateFilePath string,
) (eachResults map[string]error, criticalError error) {
	// 発注・名簿を姓で結合してマップとする
	joinedMap := make(map[Order]string, len(orders))
	for _, order := range orders {
		if fullName, exists := rosterRecords[order.FamilyName]; exists {
			joinedMap[order] = fullName
		} else {
			return eachResults, fmt.Errorf("名簿に%sさんの姓名がありません", order.FamilyName)
		}
	}

	// 発注情報についてループ
	type EachResult struct {
		fileName string
		error_   error
	}
	resultsChan := make(chan EachResult, len(joinedMap))
	var waitGroup sync.WaitGroup
	for order, fullName := range joinedMap {
		waitGroup.Add(1)
		go func(o Order, f string) {
			defer waitGroup.Done()
			fileName := o.GetOutFileName()

			workdays, err := g.workdaysRetriever.RetrieveWorkdays(
				o.StartDate.Time, o.EndDate.Time)
			if err != nil {
				resultsChan <- EachResult{
					fileName: fileName,
					error_:   err,
				}
				return
			}

			resultsChan <- EachResult{
				fileName: fileName,
				error_: g.newEachScheduleMaker(&o, workdays, f,
					filepath.Join(outputDirPath, fileName),
					templateFilePath,
				).MakeSchedule(),
			}

		}(order, fullName)
	}
	go func() {
		waitGroup.Wait()
		close(resultsChan)
	}()

	eachResults = make(map[string]error, len(joinedMap))
	for result := range resultsChan {
		eachResults[result.fileName] = result.error_
	}

	return eachResults, nil
}
