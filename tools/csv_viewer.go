package tools

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CSVViewerResult struct {
	FileName     string     `json:"file_name"`
	Headers      []string   `json:"headers"`
	Rows         [][]string `json:"rows"`
	TotalRows    int        `json:"total_rows"`
	TotalColumns int        `json:"total_columns"`
	PreviewRows  [][]string `json:"preview_rows"`
}

func HandleCSVUpload(c *gin.Context) {
	// 创建上传目录
	uploadDir := "uploads/csv"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	// 检查文件扩展名
	ext := filepath.Ext(file.Filename)
	if ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传CSV文件"})
		return
	}

	// 生成唯一文件名
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "CSV文件上传成功",
		"file":     filepath,
		"filename": file.Filename,
	})
}

func HandleCSVView(c *gin.Context) {
	filePath := c.Query("file")
	previewOnly := c.DefaultQuery("preview", "false")
	previewRows := 10 // 默认预览10行

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件路径参数"})
		return
	}

	// 读取CSV文件
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 读取所有行
	var allRows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取CSV文件失败: " + err.Error()})
			return
		}
		allRows = append(allRows, record)
	}

	if len(allRows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV文件为空"})
		return
	}

	// 获取表头
	headers := allRows[0]
	dataRows := allRows[1:]

	// 计算总行数和列数
	totalRows := len(dataRows)
	totalColumns := len(headers)

	// 准备预览数据
	var previewData [][]string
	if previewOnly == "true" {
		// 只返回预览数据
		previewData = append(previewData, headers) // 包含表头
		for i := 0; i < previewRows && i < len(dataRows); i++ {
			previewData = append(previewData, dataRows[i])
		}
	} else {
		// 返回所有数据
		previewData = allRows
	}

	result := CSVViewerResult{
		FileName:     filepath.Base(filePath),
		Headers:      headers,
		Rows:         dataRows,
		TotalRows:    totalRows,
		TotalColumns: totalColumns,
		PreviewRows:  previewData,
	}

	c.JSON(http.StatusOK, result)
}

// 获取CSV文件的统计信息
func GetCSVStats(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var allRows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		allRows = append(allRows, record)
	}

	if len(allRows) == 0 {
		return map[string]interface{}{
			"total_rows":    0,
			"total_columns": 0,
		}, nil
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	// 分析每列的数据类型
	columnTypes := make([]string, len(headers))
	for i := range headers {
		columnTypes[i] = analyzeColumnType(dataRows, i)
	}

	return map[string]interface{}{
		"total_rows":    len(dataRows),
		"total_columns": len(headers),
		"headers":       headers,
		"column_types":  columnTypes,
	}, nil
}

// 分析列的数据类型
func analyzeColumnType(rows [][]string, columnIndex int) string {
	if len(rows) == 0 {
		return "unknown"
	}

	isNumeric := true
	isInteger := true
	hasData := false

	for _, row := range rows {
		if columnIndex >= len(row) {
			continue
		}

		value := row[columnIndex]
		if value == "" {
			continue
		}

		hasData = true

		// 尝试解析为浮点数
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			isNumeric = false
			isInteger = false
			break
		}

		// 尝试解析为整数
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			isInteger = false
		}
	}

	if !hasData {
		return "empty"
	}

	if isInteger {
		return "integer"
	} else if isNumeric {
		return "float"
	} else {
		return "string"
	}
}
