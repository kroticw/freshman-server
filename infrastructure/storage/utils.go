package storage

import (
	"path"
	"strings"
)

func getFilePath(filename string, basePath string) string {
	var baseName string
	if path.Base(filename) == filename {
		baseName = filename
	} else {
		baseName = path.Dir(filename)
	}

	// Если длина названия меньше 6 символов, не получится разделить их на 3 уровня вложенности корректно.
	if len(baseName) < 3 {
		return path.Join(basePath, baseName, filename)
	}
	if len(baseName) < 4 {
		return path.Join(basePath, baseName[0:2], filename)
	}
	if len(baseName) < 6 {
		return path.Join(basePath, baseName[0:2], baseName[2:4], filename)
	}

	return path.Join(basePath, baseName[0:2], baseName[2:4], baseName[4:6], filename)
}

func getFullFileName(filename string, sourceFilename string) string {
	sourceFilename = strings.Replace(path.Base(sourceFilename), path.Ext(sourceFilename), "", -1)
	return path.Join(sourceFilename, filename)
}
