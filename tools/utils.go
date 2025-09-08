package tools

import (
	"io"
	"os"
)

// readFileContent 读取文件内容
func readFileContent(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// generateLineByLineDiff 生成逐行差异比较
func generateLineByLineDiff(lines1, lines2 []string) []DiffLine {
	var diffLines []DiffLine

	// 使用简单的LCS算法进行逐行比较
	i, j := 0, 0
	lineNum1, lineNum2 := 1, 1

	for i < len(lines1) || j < len(lines2) {
		if i >= len(lines1) {
			// 文件1已结束，剩余的都是插入
			diffLines = append(diffLines, DiffLine{
				Type:     "insert",
				Line1:    "",
				Line2:    lines2[j],
				LineNum1: 0,
				LineNum2: lineNum2,
			})
			j++
			lineNum2++
		} else if j >= len(lines2) {
			// 文件2已结束，剩余的都是删除
			diffLines = append(diffLines, DiffLine{
				Type:     "delete",
				Line1:    lines1[i],
				Line2:    "",
				LineNum1: lineNum1,
				LineNum2: 0,
			})
			i++
			lineNum1++
		} else if lines1[i] == lines2[j] {
			// 行相同
			diffLines = append(diffLines, DiffLine{
				Type:     "equal",
				Line1:    lines1[i],
				Line2:    lines2[j],
				LineNum1: lineNum1,
				LineNum2: lineNum2,
			})
			i++
			j++
			lineNum1++
			lineNum2++
		} else {
			// 行不同，需要进一步判断
			// 简单的启发式：如果下一行匹配，则当前行是修改
			if i+1 < len(lines1) && j+1 < len(lines2) && lines1[i+1] == lines2[j+1] {
				diffLines = append(diffLines, DiffLine{
					Type:     "delete",
					Line1:    lines1[i],
					Line2:    "",
					LineNum1: lineNum1,
					LineNum2: 0,
				})
				diffLines = append(diffLines, DiffLine{
					Type:     "insert",
					Line1:    "",
					Line2:    lines2[j],
					LineNum1: 0,
					LineNum2: lineNum2,
				})
				i++
				j++
				lineNum1++
				lineNum2++
			} else if i+1 < len(lines1) && lines1[i+1] == lines2[j] {
				// 文件1多了一行
				diffLines = append(diffLines, DiffLine{
					Type:     "delete",
					Line1:    lines1[i],
					Line2:    "",
					LineNum1: lineNum1,
					LineNum2: 0,
				})
				i++
				lineNum1++
			} else if j+1 < len(lines2) && lines1[i] == lines2[j+1] {
				// 文件2多了一行
				diffLines = append(diffLines, DiffLine{
					Type:     "insert",
					Line1:    "",
					Line2:    lines2[j],
					LineNum1: 0,
					LineNum2: lineNum2,
				})
				j++
				lineNum2++
			} else {
				// 都不匹配，当作修改处理
				diffLines = append(diffLines, DiffLine{
					Type:     "delete",
					Line1:    lines1[i],
					Line2:    "",
					LineNum1: lineNum1,
					LineNum2: 0,
				})
				diffLines = append(diffLines, DiffLine{
					Type:     "insert",
					Line1:    "",
					Line2:    lines2[j],
					LineNum1: 0,
					LineNum2: lineNum2,
				})
				i++
				j++
				lineNum1++
				lineNum2++
			}
		}
	}

	return diffLines
}
