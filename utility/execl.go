package utility

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strings"
)

// File excel各单元数据结构
type File struct {
	Sheets []Sheet `json:"sheets"`
}

type Sheet struct {
	Name    string   `json:"name"`    // 工作表名称
	Headers []string `json:"headers"` // 列名
	Rows    [][]any  `json:"rows"`    // 行数据（支持不同类型）
}

// CreateExcelFile 生成 Excel 文件
func CreateExcelFile(data File, fileName, filePath, host string) (string, error) {
	if err := validateFileData(data); err != nil {
		return "", err
	}
	f := excelize.NewFile()

	// 遍历 sheets
	for i, sheet := range data.Sheets {
		sheetName := sheet.Name
		if err := createSheet(f, sheet, sheetName, i); err != nil {
			return "", err
		}
	}

	// 设置默认显示的 Sheet
	f.SetActiveSheet(0)

	// 保存文件
	fullPath := filepath.Join(filePath, fileName)
	if err := ensureDirExists(filePath); err != nil {
		return "", err
	}

	if err := removeOldFile(fullPath); err != nil {
		return "", err
	}

	if err := f.SaveAs(fullPath); err != nil {
		return "", err
	}

	// 确保 host 以斜杠结尾
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	return host + fullPath, nil
}

// createSheet 创建一个工作表，并填充数据
func createSheet(f *excelize.File, sheet Sheet, sheetName string, index int) error {
	var err error
	if index == 0 {
		// 默认创建的 Sheet1 需要重命名
		err = f.SetSheetName("Sheet1", sheetName)
	} else {
		_, err = f.NewSheet(sheetName)
	}
	if err != nil {
		return err
	}

	// 写入列名
	if err = writeHeaders(f, sheetName, sheet.Headers); err != nil {
		return err
	}

	// 设置列宽
	autoAdjustColumnWidth(f, sheetName, sheet.Headers, sheet.Rows)

	// 写入数据
	if err = writeRows(f, sheetName, sheet.Rows); err != nil {
		return err
	}

	return nil
}

// writeHeaders 写入列名
func writeHeaders(f *excelize.File, sheetName string, headers []string) error {
	cellRef, _ := excelize.CoordinatesToCellName(1, 1)
	// 批量写入列名
	if err := f.SetSheetRow(sheetName, cellRef, &headers); err != nil {
		return err
	}
	return nil
}

// writeRows 批量写入数据行并合并单元格
func writeRows(f *excelize.File, sheetName string, rows [][]any) error {
	// 使用 make 创建一个指定大小的切片，初始化所有元素为 -1
	lastMergedRows := make([]int, len(rows[0]))
	for i := range lastMergedRows {
		lastMergedRows[i] = -1
	}

	// 创建居中对齐样式
	centeredStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}

	// 遍历每一行数据
	for rowIndex, row := range rows {
		cellRef, _ := excelize.CoordinatesToCellName(1, rowIndex+2) // +2 是因为第一行是列名
		if err := f.SetSheetRow(sheetName, cellRef, &row); err != nil {
			return err
		}

		// 对每一列进行检查，是否与上一行的值相同
		for colIndex := 0; colIndex < 4; colIndex++ {
			if rowIndex > 0 && row[colIndex] == rows[rowIndex-1][colIndex] {
				// 如果当前列与上一行相同，合并当前单元格与上一行的单元格
				topLeftCell := fmt.Sprintf("%s%d", string('A'+colIndex), lastMergedRows[colIndex]+2)
				bottomRightCell := fmt.Sprintf("%s%d", string('A'+colIndex), rowIndex+2)
				if err := f.MergeCell(sheetName, topLeftCell, bottomRightCell); err != nil {
					return err
				}

				// 设置合并区域的居中对齐样式
				if err := f.SetCellStyle(sheetName, topLeftCell, bottomRightCell, centeredStyle); err != nil {
					return err
				}
			} else {
				// 否则更新最后一个合并的行号
				lastMergedRows[colIndex] = rowIndex
			}

			// 对当前单元格应用居中对齐样式
			cellRef, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			if err := f.SetCellStyle(sheetName, cellRef, cellRef, centeredStyle); err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureDirExists 确保目录存在
func ensureDirExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return os.MkdirAll(filePath, 0750)
	}
	return nil
}

// removeOldFile 删除旧文件
func removeOldFile(fullPath string) error {
	if _, err := os.Stat(fullPath); err == nil {
		return os.Remove(fullPath)
	}
	return nil
}

// autoAdjustColumnWidth 自动调整列宽
func autoAdjustColumnWidth(f *excelize.File, sheetName string, headers []string, rows [][]any) {
	for colIndex, header := range headers {
		maxWidth := len(header)
		for _, row := range rows {
			cellValue := fmt.Sprintf("%v", row[colIndex])
			if len(cellValue) > maxWidth {
				maxWidth = len(cellValue)
			}
		}
		colName, _ := excelize.ColumnNumberToName(colIndex + 1)
		if err := f.SetColWidth(sheetName, colName, colName, float64(maxWidth+1)); err != nil {
			fmt.Printf("Failed to set column width for %s: %v\n", colName, err)
		}
	}
}

// validateFileData 验证文件数据
func validateFileData(data File) error {
	if len(data.Sheets) == 0 {
		return fmt.Errorf("no sheets provided")
	}
	for _, sheet := range data.Sheets {
		if len(sheet.Headers) == 0 {
			return fmt.Errorf("sheet '%s' has no headers", sheet.Name)
		}
		if len(sheet.Rows) == 0 {
			return fmt.Errorf("sheet '%s' has no rows", sheet.Name)
		}
	}
	return nil
}
