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
	// Create upload directory.
	uploadDir := "uploads/archive-compare"
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
	if ext != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please upload a ZIP archive"})
		return
	}

	// Generate unique file name.
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Extract archive.
	extractDir := filepath.Join(uploadDir, fmt.Sprintf("extracted_%d", timestamp))
	if err := extractZip(filePath, extractDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract archive: " + err.Error()})
		return
	}

	// Analyze extracted structure.
	directories, transactions, err := analyzeExtractedArchive(extractDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze archive structure: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "archive uploaded and extracted successfully",
		"archive_file": filePath,
		"extract_dir":  extractDir,
		"directories":  directories,
		"transactions": transactions,
	})
}

func HandleArchiveCompare(c *gin.Context) {
	extractDir := c.Query("extract_dir")

	if extractDir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing extract directory parameter"})
		return
	}

	// Analyze extracted structure.
	directories, transactions, err := analyzeExtractedArchive(extractDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze archive structure: " + err.Error()})
		return
	}

	// Compare trade files.
	comparisons, err := compareTransactionFiles(extractDir, transactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compare trade files: " + err.Error()})
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

	// Iterate through archive entries.
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

	// Walk the extracted directory tree.
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip root directory.
		if path == extractDir {
			return nil
		}

		// Record directory names.
		if info.IsDir() {
			relPath, _ := filepath.Rel(extractDir, path)
			if relPath != "." {
				directories = append(directories, relPath)
			}
			return nil
		}

		// Process potential trade files.
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

				// Ensure directory is tracked.
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

				// Track file metadata.
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

	// Build slice and sort.
	var transactions []TransactionInfo
	for _, t := range transactionMap {
		transactions = append(transactions, *t)
	}

	// Sort by transaction ID.
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].ID < transactions[j].ID
	})

	// Sort directory names.
	sort.Strings(directories)

	return directories, transactions, nil
}

func isTransactionFile(fileName string) bool {
	// Validate trade file names: babyy-risk-{id}.txt or candyy-risk-{id}.txt
	pattern := `^(babyy-risk-|candyy-risk-).*\.txt$`
	matched, _ := regexp.MatchString(pattern, fileName)
	return matched
}

func parseTransactionFileName(fileName string) (string, string) {
	// Example: babyy-risk-13233-2332.txt -> 13233-2332, baby
	// Example: candyy-risk-13233-2332.txt -> 13233-2332, candy

	// Remove extension.
	name := strings.TrimSuffix(fileName, ".txt")

	// Determine file type.
	if strings.HasPrefix(name, "babyy-risk-") {
		transactionID := strings.TrimPrefix(name, "babyy-risk-")
		return transactionID, "baby"
	}

	if strings.HasPrefix(name, "candyy-risk-") {
		transactionID := strings.TrimPrefix(name, "candyy-risk-")
		return transactionID, "candy"
	}

	return "", ""
}

func compareTransactionFiles(extractDir string, transactions []TransactionInfo) ([]TransactionComparison, error) {
	var comparisons []TransactionComparison

	for _, transaction := range transactions {
		// Compare files in each directory for this transaction.
		for _, dir := range transaction.Directories {
			var babyFile, candyFile string
			var babyContent, candyContent string

			// Locate baby and candy files within the directory.
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

			// Compare when both files are present.
			if babyFile != "" && candyFile != "" {
				// Generate diff.
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(babyContent, candyContent, true)
				diffHTML := dmp.DiffPrettyHtml(diffs)

				// Build line-by-line comparison.
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
