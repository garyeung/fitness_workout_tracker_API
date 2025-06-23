package util

import (
	"log"
	"path/filepath"
	"runtime"
)

func IntTo64(num int) *int64 {
	id64 := int64(num)
	return &id64

}

func GetFilePath(relativePath string) string {
	_, filename, _, ok := runtime.Caller(1)

	if !ok {
		log.Fatalf("Failed to get caller information")
	}

	dir := filepath.Dir(filename)

	return filepath.Join(dir, relativePath)
}
