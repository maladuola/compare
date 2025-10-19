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
	// Create upload directory.
	uploadDir := "uploads/csv"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	// Retrieve uploaded file.
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to retrieve uploaded file"})
		return
	}

	// Validate file extension.
	ext := filepath.Ext(file.Filename)
	if ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please upload a CSV file"})
		return
	}

	// Generate unique file name.
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "CSV file uploaded successfully",
		"file":     filepath,
		"filename": file.Filename,
	})
}

func HandleCSVView(c *gin.Context) {
	filePath := c.Query("file")
	previewOnly := c.DefaultQuery("preview", "false")
	previewRows := 10 // Default preview 10 rows.

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file path parameter"})
		return
	}

	// Read CSV file.
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file: " + err.Error()})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read all rows.
	var allRows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read CSV file: " + err.Error()})
			return
		}
		allRows = append(allRows, record)
	}

	if len(allRows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file is empty"})
		return
	}

	// Extract headers.
	headers := allRows[0]
	dataRows := allRows[1:]

	// Calculate total rows and columns.
	totalRows := len(dataRows)
	totalColumns := len(headers)

	// Prepare preview data.
	var previewData [][]string
	if previewOnly == "true" {
		// Return preview data only.
		previewData = append(previewData, headers) // Include header row.
		for i := 0; i < previewRows && i < len(dataRows); i++ {
			previewData = append(previewData, dataRows[i])
		}
	} else {
		// Return all data.
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

// GetCSVStats returns summary information for a CSV file.
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

	// Determine column data types.
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

// analyzeColumnType inspects a column and returns its inferred type.
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

		// Try parsing as float.
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			isNumeric = false
			isInteger = false
			break
		}

		// Try parsing as integer.
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
