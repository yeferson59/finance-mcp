package file

import (
	"path/filepath"
	"runtime"
)

func GetPathFile(fileName string) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	envPath := filepath.Join(projectRoot, fileName)
	return envPath
}
