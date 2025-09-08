package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type ArchiveCompareResult struct {
	ArchiveName  string                  `json:"archive_name"`
	Directories  []string                `json:"directories"`
	Transactions []TransactionInfo       `json:"transactions"`
	Comparisons  []TransactionComparison `json:"comparisons"`
}

type TransactionInfo struct {
	ID          string            `json:"id"`
	Directories []string          `json:"directories"`
	Files       []TransactionFile `json:"files"`
}

type TransactionFile struct {
	Directory string `json:"directory"`
	FileName  string `json:"file_name"`
	FilePath  string `json:"file_path"`
	Type      string `json:"type"` // "baby" or "candy"
}

type TransactionComparison struct {
	TransactionID string     `json:"transaction_id"`
	Directory     string     `json:"directory"`
	BabyFile      string     `json:"baby_file"`
	CandyFile     string     `json:"candy_file"`
	BabyContent   string     `json:"baby_content"`
	CandyContent  string     `json:"candy_content"`
	DiffHTML      string     `json:"diff_html"`
	DiffLines     []DiffLine `json:"diff_lines"`
}

func HandleArchiveUpload(c *gin.Context) {
	// 创建上传目录
	uploadDir := "uploads/archive-compare"
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
	if ext != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传ZIP压缩文件"})
		return
	}

	// 生成唯一文件名
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 解压文件
	extractDir := filepath.Join(uploadDir, fmt.Sprintf("extracted_%d", timestamp))
	if err := extractZip(filePath, extractDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解压文件失败: " + err.Error()})
		return
	}

	// 分析解压后的目录结构
	directories, transactions, err := analyzeExtractedArchive(extractDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分析文件结构失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "压缩文件上传和解压成功",
		"archive_file": filePath,
		"extract_dir":  extractDir,
		"directories":  directories,
		"transactions": transactions,
	})
}

func HandleArchiveCompare(c *gin.Context) {
	extractDir := c.Query("extract_dir")

	if extractDir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少解压目录参数"})
		return
	}

	// 分析解压后的目录结构
	directories, transactions, err := analyzeExtractedArchive(extractDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分析文件结构失败: " + err.Error()})
		return
	}

	// 进行交易文件比较
	comparisons, err := compareTransactionFiles(extractDir, transactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "比较交易文件失败: " + err.Error()})
		return
	}

	result := ArchiveCompareResult{
		ArchiveName:  filepath.Base(extractDir),
		Directories:  directories,
		Transactions: transactions,
		Comparisons:  comparisons,
	}

	c.JSON(http.StatusOK, result)
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	// 遍历zip文件中的文件
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
			if err != nil {
				rc.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func analyzeExtractedArchive(extractDir string) ([]string, []TransactionInfo, error) {
	var directories []string
	transactionMap := make(map[string]*TransactionInfo)

	// 遍历解压后的目录
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if path == extractDir {
			return nil
		}

		// 如果是目录，记录目录名
		if info.IsDir() {
			relPath, _ := filepath.Rel(extractDir, path)
			if relPath != "." {
				directories = append(directories, relPath)
			}
			return nil
		}

		// 如果是文件，检查是否是交易文件
		fileName := info.Name()
		if isTransactionFile(fileName) {
			transactionID, fileType := parseTransactionFileName(fileName)
			if transactionID != "" {
				relPath, _ := filepath.Rel(extractDir, path)
				dir := filepath.Dir(relPath)

				if transactionMap[transactionID] == nil {
					transactionMap[transactionID] = &TransactionInfo{
						ID:          transactionID,
						Directories: []string{},
						Files:       []TransactionFile{},
					}
				}

				// 添加目录（如果还没有）
				hasDir := false
				for _, d := range transactionMap[transactionID].Directories {
					if d == dir {
						hasDir = true
						break
					}
				}
				if !hasDir {
					transactionMap[transactionID].Directories = append(transactionMap[transactionID].Directories, dir)
				}

				// 添加文件
				transactionMap[transactionID].Files = append(transactionMap[transactionID].Files, TransactionFile{
					Directory: dir,
					FileName:  fileName,
					FilePath:  path,
					Type:      fileType,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// 转换为切片并排序
	var transactions []TransactionInfo
	for _, t := range transactionMap {
		transactions = append(transactions, *t)
	}

	// 按交易ID排序
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].ID < transactions[j].ID
	})

	// 目录排序
	sort.Strings(directories)

	return directories, transactions, nil
}

func isTransactionFile(fileName string) bool {
	// 检查文件名是否符合模式：babyy-risk-{id}.txt 或 candyy-risk-{id}.txt
	pattern := `^(babyy-risk-|candyy-risk-).*\.txt$`
	matched, _ := regexp.MatchString(pattern, fileName)
	return matched
}

func parseTransactionFileName(fileName string) (string, string) {
	// 解析文件名：babyy-risk-13233-2332.txt -> 13233-2332, baby
	// 解析文件名：candyy-risk-13233-2332.txt -> 13233-2332, candy

	// 移除.txt扩展名
	name := strings.TrimSuffix(fileName, ".txt")

	// 检查是否是baby文件
	if strings.HasPrefix(name, "babyy-risk-") {
		transactionID := strings.TrimPrefix(name, "babyy-risk-")
		return transactionID, "baby"
	}

	// 检查是否是candy文件
	if strings.HasPrefix(name, "candyy-risk-") {
		transactionID := strings.TrimPrefix(name, "candyy-risk-")
		return transactionID, "candy"
	}

	return "", ""
}

func compareTransactionFiles(extractDir string, transactions []TransactionInfo) ([]TransactionComparison, error) {
	var comparisons []TransactionComparison

	for _, transaction := range transactions {
		// 为每个目录比较该交易的文件
		for _, dir := range transaction.Directories {
			var babyFile, candyFile string
			var babyContent, candyContent string

			// 查找该目录下的baby和candy文件
			for _, file := range transaction.Files {
				if file.Directory == dir {
					content, err := readFileContent(file.FilePath)
					if err != nil {
						continue
					}

					if file.Type == "baby" {
						babyFile = file.FileName
						babyContent = content
					} else if file.Type == "candy" {
						candyFile = file.FileName
						candyContent = content
					}
				}
			}

			// 如果找到了两个文件，进行比较
			if babyFile != "" && candyFile != "" {
				// 生成差异
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(babyContent, candyContent, true)
				diffHTML := dmp.DiffPrettyHtml(diffs)

				// 生成逐行比较
				lines1 := strings.Split(babyContent, "\n")
				lines2 := strings.Split(candyContent, "\n")
				diffLines := generateLineByLineDiff(lines1, lines2)

				comparison := TransactionComparison{
					TransactionID: transaction.ID,
					Directory:     dir,
					BabyFile:      babyFile,
					CandyFile:     candyFile,
					BabyContent:   babyContent,
					CandyContent:  candyContent,
					DiffHTML:      diffHTML,
					DiffLines:     diffLines,
				}

				comparisons = append(comparisons, comparison)
			}
		}
	}

	return comparisons, nil
}
