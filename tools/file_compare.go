package tools

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type FileCompareResult struct {
	File1Name    string     `json:"file1_name"`
	File2Name    string     `json:"file2_name"`
	File1Content string     `json:"file1_content"`
	File2Content string     `json:"file2_content"`
	DiffHTML     string     `json:"diff_html"`
	Lines1       []string   `json:"lines1"`
	Lines2       []string   `json:"lines2"`
	DiffLines    []DiffLine `json:"diff_lines"`
}

func HandleFileCompareUpload(c *gin.Context) {
	// 创建上传目录
	uploadDir := "uploads/file-compare"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	files := form.File["files"]
	if len(files) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传两个文件进行比较"})
		return
	}

	// 保存文件
	var savedFiles []string
	for i, file := range files {
		// 生成唯一文件名
		timestamp := time.Now().UnixNano()
		filename := fmt.Sprintf("%d_%s", timestamp+int64(i), file.Filename)
		filepath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filepath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}
		savedFiles = append(savedFiles, filepath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文件上传成功",
		"files":   savedFiles,
	})
}

func HandleFileCompare(c *gin.Context) {
	file1Path := c.Query("file1")
	file2Path := c.Query("file2")

	if file1Path == "" || file2Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件路径参数"})
		return
	}

	// 读取文件内容
	content1, err := readFileContent(file1Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件1失败: " + err.Error()})
		return
	}

	content2, err := readFileContent(file2Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件2失败: " + err.Error()})
		return
	}

	// 生成差异
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(content1, content2, true)

	// 生成HTML格式的差异
	diffHTML := dmp.DiffPrettyHtml(diffs)

	// 生成逐行比较
	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")
	diffLines := generateLineByLineDiff(lines1, lines2)

	result := FileCompareResult{
		File1Name:    filepath.Base(file1Path),
		File2Name:    filepath.Base(file2Path),
		File1Content: content1,
		File2Content: content2,
		DiffHTML:     diffHTML,
		Lines1:       lines1,
		Lines2:       lines2,
		DiffLines:    diffLines,
	}

	c.JSON(http.StatusOK, result)
}
