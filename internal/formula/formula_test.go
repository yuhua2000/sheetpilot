package formula

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func newTestFile(t *testing.T) *excelize.File {
	t.Helper()
	f := excelize.NewFile()
	t.Cleanup(func() { f.Close() })
	return f
}

func TestGetFormula(t *testing.T) {
	f := newTestFile(t)

	// Set a formula
	f.SetCellFormula("Sheet1", "C1", "=A1+B1")

	formula, err := GetFormula(f, "Sheet1", "C1")
	require.NoError(t, err)
	require.Equal(t, "=A1+B1", formula)
}

func TestGetFormula_NoFormula(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Hello")

	formula, err := GetFormula(f, "Sheet1", "A1")
	require.NoError(t, err)
	require.Empty(t, formula)
}

func TestSetFormula(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", 10)
	f.SetCellValue("Sheet1", "B1", 20)

	err := SetFormula(f, "Sheet1", "C1", "=A1+B1")
	require.NoError(t, err)

	formula, _ := f.GetCellFormula("Sheet1", "C1")
	require.Equal(t, "=A1+B1", formula)
}

func TestFillFormulaColumn(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Price")
	f.SetCellValue("Sheet1", "B1", "Qty")
	f.SetCellValue("Sheet1", "C1", "Total")
	f.SetCellValue("Sheet1", "A2", 10)
	f.SetCellValue("Sheet1", "B2", 5)
	f.SetCellValue("Sheet1", "A3", 20)
	f.SetCellValue("Sheet1", "B3", 3)

	err := FillFormulaColumn(f, "Sheet1", "C", 2, 3, "=A{row}*B{row}")
	require.NoError(t, err)

	// Verify formulas were set
	formula2, _ := f.GetCellFormula("Sheet1", "C2")
	require.NotEmpty(t, formula2)

	formula3, _ := f.GetCellFormula("Sheet1", "C3")
	require.NotEmpty(t, formula3)
}
