package utils

import (
	"archive/zip"
	"strings"
)

// CountPptxSlides 统计 PPTX 文件的幻灯片页数
// PPTX 本质是 ZIP 包，幻灯片存储在 ppt/slides/slideN.xml
func CountPptxSlides(filePath string) (int, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	count := 0
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			count++
		}
	}
	return count, nil
}
