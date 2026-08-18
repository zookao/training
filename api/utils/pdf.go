package utils

import (
	"os"
	"regexp"
)

// pdfPageRegexp 匹配 PDF 中的 /Type /Page 对象（排除 /Type /Pages 父节点）
// LibreOffice 导出的 PDF 结构未压缩，此正则可可靠计数
var pdfPageRegexp = regexp.MustCompile(`/Type\s*/Page[^s]`)

// CountPDFPages 统计 PDF 文件的页数
// 通过匹配 /Type /Page（非 /Pages）对象计数，适用于 LibreOffice 生成的 PDF
func CountPDFPages(pdfPath string) (int, error) {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return 0, err
	}
	return len(pdfPageRegexp.FindAll(data, -1)), nil
}
