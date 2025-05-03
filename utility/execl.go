package utility

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"log"
)

// File excel 各单元数据结构
type File struct {
	Sheets []Sheet `json:"sheets"`
}

type Sheet struct {
	Name    string   `json:"name"`    // 工作表名称
	Headers []string `json:"headers"` // 列名
	Rows    [][]any  `json:"rows"`    // 行数据（支持不同类型）
}

// Constant definitions
const (
	DefaultSheetName  = "Sheet1"
	DefaultDirPerm    = 0750
	DefaultHostSuffix = "/"
	MinColumnWidth    = 10 // 最小列宽
	MergeCheckColumns = 4  // 检查是否需要合并单元格的列数
)

// CreateExcelFile 生成 Excel 文件
func CreateExcelFile(data File, fileName, filePath, host string) (string, error) {
	// 验证数据
	if err := validateFileData(data); err != nil {
		return "", fmt.Errorf("无效的文件数据: %w", err)
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("关闭 Excel 文件失败: %v", err)
		}
	}()

	// 处理每个工作表
	for i, sheet := range data.Sheets {
		if err := createSheet(f, sheet, i); err != nil {
			return "", fmt.Errorf("创建工作表 '%s' 失败: %w", sheet.Name, err)
		}
	}

	// 设置默认显示的工作表
	f.SetActiveSheet(0)

	// 保存文件
	fullPath := filepath.Join(filePath, fileName)
	if err := ensureDirExists(filePath); err != nil {
		return "", fmt.Errorf("确保目录存在失败: %w", err)
	}

	if err := removeOldFile(fullPath); err != nil {
		log.Printf("删除旧文件失败: %v", err) // 非关键错误，记录警告
	}

	if err := f.SaveAs(fullPath); err != nil {
		return "", fmt.Errorf("保存 Excel 文件失败: %w", err)
	}

	// 确保主机地址以斜杠结尾
	if !strings.HasSuffix(host, DefaultHostSuffix) {
		host += DefaultHostSuffix
	}

	return host + fullPath, nil
}

// createSheet 创建一个工作表，并填充数据
func createSheet(f *excelize.File, sheet Sheet, index int) error {
	sheetName := sheet.Name
	if index == 0 {
		// 重命名默认的 Sheet1
		if err := f.SetSheetName(DefaultSheetName, sheetName); err != nil {
			return fmt.Errorf("重命名默认工作表失败: %w", err)
		}
	} else {
		if _, err := f.NewSheet(sheetName); err != nil {
			return fmt.Errorf("创建新工作表失败: %w", err)
		}
	}

	// 创建流式写入器
	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("创建流式写入器失败: %w", err)
	}
	defer func() {
		if err := sw.Flush(); err != nil {
			log.Printf("刷新流式写入器失败: %v", err)
		}
	}()

	// 设置列宽
	autoAdjustColumnWidth(sw, sheet.Headers, sheet.Rows)

	// 创建加粗样式的列名
	headerStyleID, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return fmt.Errorf("创建列名样式失败: %w", err)
	}

	// 创建居中对齐样式
	centeredStyleID, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return fmt.Errorf("创建居中对齐样式失败: %w", err)
	}

	// 写入列名
	if err := writeHeaders(sw, sheet.Headers, headerStyleID); err != nil {
		return fmt.Errorf("写入列名失败: %w", err)
	}

	// 写入数据行
	if err := writeRows(sw, sheet.Rows, centeredStyleID); err != nil {
		return fmt.Errorf("写入数据行失败: %w", err)
	}

	return nil
}

// writeHeaders 写入列名到工作表
func writeHeaders(sw *excelize.StreamWriter, headers []string, styleID int) error {
	cellRef, _ := excelize.CoordinatesToCellName(1, 1)

	// 将列名转换为 []any
	headerInterfaces := make([]any, len(headers))
	for i, header := range headers {
		headerInterfaces[i] = header
	}

	if err := sw.SetRow(cellRef, headerInterfaces, excelize.RowOpts{StyleID: styleID}); err != nil {
		return fmt.Errorf("设置列名行失败: %w", err)
	}

	return nil
}

// writeRows 写入数据行到工作表，并在需要时合并单元格
func writeRows(sw *excelize.StreamWriter, rows [][]any, centeredStyle int) error {
	numCols := len(rows[0])
	lastMergedRows := make([]int, MergeCheckColumns)
	for i := range lastMergedRows {
		lastMergedRows[i] = -1
	}

	// 预分配行数据切片，避免重复分配
	rowInterfaces := make([]interface{}, numCols)

	// 遍历每一行数据
	for rowIndex, row := range rows {
		cellRef, _ := excelize.CoordinatesToCellName(1, rowIndex+2) // +2 是因为第一行是列名，且索引从 1 开始

		for colIndex, cell := range row {
			c := excelize.Cell{Value: cell}
			if colIndex < MergeCheckColumns { // 对前 MergeCheckColumns 列应用居中对齐样式
				c.StyleID = centeredStyle
			}
			rowInterfaces[colIndex] = c
		}

		if err := sw.SetRow(cellRef, rowInterfaces); err != nil {
			return fmt.Errorf("设置数据行失败: %w", err)
		}

		// 检查是否需要在前 MergeCheckColumns 列中合并单元格
		for colIndex := 0; colIndex < MergeCheckColumns; colIndex++ {
			if rowIndex > 0 && row[colIndex] == rows[rowIndex-1][colIndex] {
				// 合并当前单元格与上一行的单元格
				topLeftCell := fmt.Sprintf("%s%d", string('A'+colIndex), lastMergedRows[colIndex]+2)
				bottomRightCell := fmt.Sprintf("%s%d", string('A'+colIndex), rowIndex+2)
				if err := sw.MergeCell(topLeftCell, bottomRightCell); err != nil {
					return fmt.Errorf("合并单元格失败: %w", err)
				}
			} else {
				// 更新最后一个合并的行号
				lastMergedRows[colIndex] = rowIndex
			}
		}
	}

	return nil
}

// ensureDirExists 确保目录存在
func ensureDirExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filePath, DefaultDirPerm); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return nil
}

// removeOldFile 删除旧文件
func removeOldFile(fullPath string) error {
	if _, err := os.Stat(fullPath); err == nil {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("删除旧文件失败: %w", err)
		}
	}
	return nil
}

// autoAdjustColumnWidth 自动调整列宽
func autoAdjustColumnWidth(sw *excelize.StreamWriter, headers []string, rows [][]any) {
	columnWidths := make([]int, len(headers))
	for colIndex, header := range headers {
		maxWidth := len(header)
		for _, row := range rows {
			cellValue := fmt.Sprintf("%v", row[colIndex])
			if len(cellValue) > maxWidth {
				maxWidth = len(cellValue)
			}
		}
		// 确保最小列宽
		if maxWidth < MinColumnWidth {
			maxWidth = MinColumnWidth
		}
		columnWidths[colIndex] = maxWidth
	}

	// 一次性设置所有列宽
	for colIndex, width := range columnWidths {
		if err := sw.SetColWidth(colIndex+1, colIndex+1, float64(width+1)); err != nil {
			log.Printf("设置列宽失败，列索引: %d, 错误: %v", colIndex+1, err) // 记录错误，但不返回
		}
	}
}

// validateFileData 验证文件数据
func validateFileData(data File) error {
	if len(data.Sheets) == 0 {
		return fmt.Errorf("没有提供工作表")
	}
	for _, sheet := range data.Sheets {
		if len(sheet.Name) == 0 {
			return fmt.Errorf("工作表名称不能为空")
		}
		if len(sheet.Headers) == 0 {
			return fmt.Errorf("工作表 '%s' 没有列名", sheet.Name)
		}
		if len(sheet.Rows) == 0 {
			return fmt.Errorf("工作表 '%s' 没有数据行", sheet.Name)
		}
		// 检查所有行是否具有与列名相同的列数
		headerLen := len(sheet.Headers) // 缓存列名长度
		for _, row := range sheet.Rows {
			if len(row) != headerLen {
				return fmt.Errorf("工作表 '%s' 行长度与列名长度不匹配", sheet.Name)
			}
		}
	}
	return nil
}
