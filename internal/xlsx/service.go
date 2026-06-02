package xlsx

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Service struct{}

type WorkbookInfo struct {
	Sheets []string `json:"sheets"`
}

func (Service) Inspect(path string) (WorkbookInfo, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return WorkbookInfo{}, err
	}
	defer f.Close()
	return WorkbookInfo{Sheets: f.GetSheetList()}, nil
}

func (Service) ReadRange(path, sheet, cellRange string) ([][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if strings.TrimSpace(cellRange) == "" {
		return f.GetRows(sheet)
	}
	start, end, err := splitRange(cellRange)
	if err != nil {
		return nil, err
	}
	startCol, startRow, err := excelize.CellNameToCoordinates(start)
	if err != nil {
		return nil, err
	}
	endCol, endRow, err := excelize.CellNameToCoordinates(end)
	if err != nil {
		return nil, err
	}
	if startCol > endCol || startRow > endRow {
		return nil, fmt.Errorf("invalid range %q", cellRange)
	}
	out := make([][]string, 0, endRow-startRow+1)
	for r := startRow; r <= endRow; r++ {
		row := make([]string, 0, endCol-startCol+1)
		for c := startCol; c <= endCol; c++ {
			cell, err := excelize.CoordinatesToCellName(c, r)
			if err != nil {
				return nil, err
			}
			value, err := f.GetCellValue(sheet, cell)
			if err != nil {
				return nil, err
			}
			row = append(row, value)
		}
		out = append(out, row)
	}
	return out, nil
}

func (Service) WriteCell(inputPath, outputPath, sheet, cell string, value any) error {
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.SetCellValue(sheet, cell, value); err != nil {
		return err
	}
	return f.SaveAs(outputPath)
}

func (Service) AddSheet(inputPath, outputPath, sheet string) error {
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	return f.SaveAs(outputPath)
}

func splitRange(cellRange string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(cellRange), ":")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid range %q", cellRange)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
