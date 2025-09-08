package tools

// DiffLine 表示差异行
type DiffLine struct {
	Type     string `json:"type"` // "equal", "delete", "insert"
	Line1    string `json:"line1"`
	Line2    string `json:"line2"`
	LineNum1 int    `json:"line_num1"`
	LineNum2 int    `json:"line_num2"`
}
