package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"mcp-claude-tools/internal/security"
)

type FileTool struct {
	BaseDir string
}

func NewFileTool(baseDir string) *FileTool {
	return &FileTool{BaseDir: baseDir}
}

// Baca membaca file dengan dukungan offset baris dan limit baris
func (ft *FileTool) Baca(targetPath string, offsetLines, limitLines int) (string, error) {
	cleanPath, err := security.ValidatePath(ft.BaseDir, targetPath)
	if err != nil {
		return "", err
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !security.IsSizeSafe(stat.Size()) {
		return "", fmt.Errorf("error: ukuran file melebihi batas maksimum 10MB")
	}

	scanner := bufio.NewScanner(file)
	var lines []string
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine <= offsetLines {
			continue
		}
		lines = append(lines, scanner.Text())
		if limitLines > 0 && len(lines) >= limitLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

// Tulis menulis data teks langsung ke file (overwrite/create)
func (ft *FileTool) Tulis(targetPath string, content string) (string, error) {
	cleanPath, err := security.ValidatePath(ft.BaseDir, targetPath)
	if err != nil {
		return "", err
	}

	// Buat direktori pembungkus jika belum ada secara aman
	err = os.MkdirAll(filepath.Dir(cleanPath), 0755)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(cleanPath, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Berhasil menulis ke %s", cleanPath), nil
}

// Sunting melakukan penggantian kata/string persis di dalam file
func (ft *FileTool) Sunting(targetPath string, oldString, newString string) (string, error) {
	cleanPath, err := security.ValidatePath(ft.BaseDir, targetPath)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", err
	}

	content := string(data)
	if !strings.Contains(content, oldString) {
		return "", fmt.Errorf("error: teks asal '%s' tidak ditemukan di dalam file", oldString)
	}

	// Lakukan replacement persis
	updatedContent := strings.ReplaceAll(content, oldString, newString)

	err = os.WriteFile(cleanPath, []byte(updatedContent), 0644)
	if err != nil {
		return "", err
	}

	return "File berhasil diperbarui dengan perubahan teks", nil
}

// Glob mencari pola file/folder di dalam base directory
func (ft *FileTool) Glob(pattern string) ([]string, error) {
	// Pola pencarian dipaksa relatif terhadap BaseDir demi keamanan
	fullPattern := filepath.Join(ft.BaseDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}

	var relativeMatches []string
	for _, match := range matches {
		rel, err := filepath.Rel(ft.BaseDir, match)
		if err == nil {
			relativeMatches = append(relativeMatches, rel)
		}
	}

	return relativeMatches, nil
}

// Grep mencari teks berbasis regex menggunakan utilitas ripgrep (rg)
func (ft *FileTool) Grep(regexPattern string) (string, error) {
	// Memastikan ripgrep terinstal di sistem server
	_, err := exec.LookPath("rg")
	if err != nil {
		return "", fmt.Errorf("error: ripgrep ('rg') tidak terinstal di server host")
	}

	// Jalankan ripgrep di dalam BaseDir agar aman
	cmd := exec.Command("rg", "--no-heading", "--line-number", regexPattern, ft.BaseDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run() // ripgrep mengembalikan exit status 1 jika tidak ada kecocokan

	if stderr.Len() > 0 {
		return "", fmt.Errorf("ripgrep error: %s", stderr.String())
	}

	return stdout.String(), nil
}
