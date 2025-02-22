package utils

import (
	"os"
	"sort"
)

// 列出指定文件夹的一级子文件夹,并按时间先后顺序排序
func ListImmediateSubDirectories(dirPath string) ([]os.DirEntry, error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var subDirectories []os.DirEntry

	for _, entry := range dirEntries {
		if entry.IsDir() {
			subDirectories = append(subDirectories, entry)
		}
	}

	// 按照修改时间排序
	sort.Slice(subDirectories, func(i, j int) bool {
		infoI, _ := subDirectories[i].Info()
		infoJ, _ := subDirectories[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return subDirectories, nil
}
