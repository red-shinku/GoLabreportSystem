package service

import (
	"LabSystem/internal/domain"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// sheetParser 表格解析器接口
// 业务层依赖该接口而非具体格式实现，便于未来接入CSV等其他来源
type sheetParser interface {
	Parse(r io.Reader) (*domain.SheetData, error)
}

// ExcelSheetParser 基于 excelize 的 xlsx 表格解析器
// 约定的表头：第1列 ID、第2列 SNO、第3列 Sname、第4列起为各项目名
type ExcelSheetParser struct{}

func NewExcelSheetParser() *ExcelSheetParser { return &ExcelSheetParser{} }

func (p *ExcelSheetParser) Parse(r io.Reader) (*domain.SheetData, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("ExcelSheetParser.Parse() open: %w: %v", domain.ErrSheetFormat, err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("ExcelSheetParser.Parse(): %w: no sheet", domain.ErrSheetFormat)
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("ExcelSheetParser.Parse() read rows: %w: %v", domain.ErrSheetFormat, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("ExcelSheetParser.Parse(): %w: empty sheet", domain.ErrSheetFormat)
	}

	header := rows[0]
	if len(header) < 4 {
		return nil, fmt.Errorf("ExcelSheetParser.Parse(): %w: header needs >=4 columns", domain.ErrSheetFormat)
	}
	if !strings.EqualFold(strings.TrimSpace(header[0]), "ID") ||
		!strings.EqualFold(strings.TrimSpace(header[1]), "SNO") ||
		!strings.EqualFold(strings.TrimSpace(header[2]), "Sname") {
		return nil, fmt.Errorf("ExcelSheetParser.Parse(): %w: expect [ID SNO Sname ...]", domain.ErrSheetFormat)
	}

	var projectNames []string
	for _, name := range header[3:] {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		projectNames = append(projectNames, name)
	}
	if len(projectNames) == 0 {
		return nil, fmt.Errorf("ExcelSheetParser.Parse(): %w: no project column", domain.ErrSheetFormat)
	}

	var students []domain.StudentRow
	for _, row := range rows[1:] {
		if len(row) < 3 {
			continue
		}
		number := strings.TrimSpace(row[1])
		name := strings.TrimSpace(row[2])
		if number == "" || name == "" {
			continue
		}
		students = append(students, domain.StudentRow{Number: number, Name: name})
	}

	return domain.NewSheetData(students, projectNames), nil
}
