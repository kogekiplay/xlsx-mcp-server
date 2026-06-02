package xlsx

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	maxReadRows          = 500
	maxReadColumns       = 50
	maxReadCells         = 5000
	maxWorkbookUnzipSize = 64 << 20
	maxWorkbookXMLSize   = 16 << 20
)

type Service struct{}

type WorkbookInfo struct {
	Sheets []string `json:"sheets"`
}

func openWorkbook(path string) (*excelize.File, error) {
	return excelize.OpenFile(path, excelize.Options{
		UnzipSizeLimit:    maxWorkbookUnzipSize,
		UnzipXMLSizeLimit: maxWorkbookXMLSize,
	})
}

func (Service) Inspect(path string) (WorkbookInfo, error) {
	f, err := openWorkbook(path)
	if err != nil {
		return WorkbookInfo{}, err
	}
	defer f.Close()
	return WorkbookInfo{Sheets: f.GetSheetList()}, nil
}

func (Service) ReadRange(path, sheet, cellRange string) ([][]string, error) {
	f, err := openWorkbook(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if strings.TrimSpace(cellRange) == "" {
		return readRows(f, sheet)
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
	rows := endRow - startRow + 1
	cols := endCol - startCol + 1
	if err := checkReadSize(rows, cols); err != nil {
		return nil, err
	}
	out := make([][]string, 0, rows)
	for r := startRow; r <= endRow; r++ {
		row := make([]string, 0, cols)
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
	f, err := openWorkbook(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.SetCellValue(sheet, cell, value); err != nil {
		return err
	}
	if err := f.UpdateLinkedValue(); err != nil {
		return err
	}
	return f.SaveAs(outputPath)
}

func (Service) AddSheet(inputPath, outputPath, sheet string) error {
	f, err := openWorkbook(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	return f.SaveAs(outputPath)
}

func readRows(f *excelize.File, sheet string) ([][]string, error) {
	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([][]string, 0)
	cellCount := 0
	for rows.Next() {
		if len(out) >= maxReadRows {
			return nil, readSizeError()
		}
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		if len(cols) > maxReadColumns {
			return nil, readSizeError()
		}
		cellCount += len(cols)
		if cellCount > maxReadCells {
			return nil, readSizeError()
		}
		out = append(out, cols)
	}
	if err := rows.Error(); err != nil {
		return nil, err
	}
	return out, nil
}

func splitRange(cellRange string) (string, string, error) {
	trimmed := strings.TrimSpace(cellRange)
	parts := strings.Split(trimmed, ":")
	switch len(parts) {
	case 1:
		cell := strings.TrimSpace(parts[0])
		if cell == "" {
			return "", "", fmt.Errorf("invalid range %q", cellRange)
		}
		return cell, cell, nil
	case 2:
		start := strings.TrimSpace(parts[0])
		end := strings.TrimSpace(parts[1])
		if start == "" || end == "" {
			return "", "", fmt.Errorf("invalid range %q", cellRange)
		}
		return start, end, nil
	default:
		return "", "", fmt.Errorf("invalid range %q", cellRange)
	}
}

func checkReadSize(rows, columns int) error {
	if rows > maxReadRows || columns > maxReadColumns || rows*columns > maxReadCells {
		return readSizeError()
	}
	return nil
}

func readSizeError() error {
	return fmt.Errorf("read range exceeds limit of %d rows, %d columns, or %d cells", maxReadRows, maxReadColumns, maxReadCells)
}
