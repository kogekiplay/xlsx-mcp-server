package xlsx

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func makeWorkbook(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetCellValue("Sheet1", "A1", "Name"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "B1", "Amount"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "A2", "Alice"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "B2", 42); err != nil {
		t.Fatal(err)
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

func TestInspectWorkbook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	makeWorkbook(t, path)

	svc := Service{}
	info, err := svc.Inspect(path)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if len(info.Sheets) != 1 || info.Sheets[0] != "Sheet1" {
		t.Fatalf("Sheets = %#v", info.Sheets)
	}
}

func TestReadRange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	makeWorkbook(t, path)

	svc := Service{}
	values, err := svc.ReadRange(path, "Sheet1", "A1:B2")
	if err != nil {
		t.Fatalf("ReadRange() error = %v", err)
	}
	if values[1][0] != "Alice" || values[1][1] != "42" {
		t.Fatalf("values = %#v", values)
	}
}

func TestReadSingleCellRange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	makeWorkbook(t, path)

	svc := Service{}
	values, err := svc.ReadRange(path, "Sheet1", "B2")
	if err != nil {
		t.Fatalf("ReadRange() error = %v", err)
	}
	if len(values) != 1 || len(values[0]) != 1 || values[0][0] != "42" {
		t.Fatalf("values = %#v", values)
	}
}

func TestReadRangeRejectsOversizedRange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	makeWorkbook(t, path)

	svc := Service{}
	_, err := svc.ReadRange(path, "Sheet1", "A1:AZ200")
	if err == nil || !strings.Contains(err.Error(), "exceeds limit") {
		t.Fatalf("ReadRange() err = %v", err)
	}
}

func TestReadAllRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	makeWorkbook(t, path)

	svc := Service{}
	values, err := svc.ReadRange(path, "Sheet1", "")
	if err != nil {
		t.Fatalf("ReadRange() error = %v", err)
	}
	if len(values) != 2 || values[0][0] != "Name" || values[1][0] != "Alice" {
		t.Fatalf("values = %#v", values)
	}
}

func TestReadAllRowsRejectsOversizedSheet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	f := excelize.NewFile()
	defer f.Close()
	for row := 1; row <= maxReadRows+1; row++ {
		cell, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.SetCellValue("Sheet1", cell, row); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	svc := Service{}
	_, err := svc.ReadRange(path, "Sheet1", "")
	if err == nil || !strings.Contains(err.Error(), "exceeds limit") {
		t.Fatalf("ReadRange() err = %v", err)
	}
}

func TestWriteCellAndSaveAs(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "book.xlsx")
	output := filepath.Join(dir, "out.xlsx")
	makeWorkbook(t, input)

	svc := Service{}
	if err := svc.WriteCell(input, output, "Sheet1", "C1", "Status"); err != nil {
		t.Fatalf("WriteCell() error = %v", err)
	}
	f, err := excelize.OpenFile(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	value, err := f.GetCellValue("Sheet1", "C1")
	if err != nil {
		t.Fatal(err)
	}
	if value != "Status" {
		t.Fatalf("C1 = %q", value)
	}
}

func TestAddSheet(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "book.xlsx")
	output := filepath.Join(dir, "out.xlsx")
	makeWorkbook(t, input)

	svc := Service{}
	if err := svc.AddSheet(input, output, "Summary"); err != nil {
		t.Fatalf("AddSheet() error = %v", err)
	}
	f, err := excelize.OpenFile(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if index, err := f.GetSheetIndex("Summary"); err != nil || index < 0 {
		t.Fatalf("Summary sheet missing, index=%d err=%v", index, err)
	}
}
