package tools

import (
	"io"
	"os"
)

// readFileContent reads the entire file content into a string.
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

// generateLineByLineDiff builds a line-by-line diff result.
func generateLineByLineDiff(lines1, lines2 []string) []DiffLine {
	var diffLines []DiffLine

	// Use a simple LCS-inspired comparison.
	i, j := 0, 0
	lineNum1, lineNum2 := 1, 1

	for i < len(lines1) || j < len(lines2) {
		if i >= len(lines1) {
			// File1 is exhausted, remaining lines are inserts.
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
			// File2 is exhausted, remaining lines are deletes.
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
			// Lines are identical.
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
			// Lines differ, apply heuristics.
			// Simple heuristic: if the next line matches, treat current line as a modification.
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
				// File1 has an extra line.
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
				// File2 has an extra line.
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
				// Nothing matches, treat as a modification.
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
