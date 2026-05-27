package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const MaxFileSize = 10 * 1024 * 1024 // 10MB

// ValidatePath memastikan jalur yang diakses adalah absolut dan aman
func ValidatePath(baseDir, targetPath string) (string, error) {
	if !filepath.IsAbs(targetPath) {
		return "", errors.New("error: hanya menerima jalur absolut (absolute path)")
	}

	cleanTarget := filepath.Clean(targetPath)
	cleanBase := filepath.Clean(baseDir)

	// Memastikan target berada di dalam base directory untuk mencegah traversal
	if !strings.HasPrefix(cleanTarget, cleanBase) {
		return "", fmt.Errorf("error: akses ditolak, jalur di luar direktori kerja yang diizinkan")
	}

	return cleanTarget, nil
}

// IsSizeSafe memeriksa apakah ukuran file tidak melebihi batas
func IsSizeSafe(size int64) bool {
	return size <= MaxFileSize
}
