package utils

import (
    "errors"
    "path/filepath"
    "strings"
)

var (
    ErrInvalidFileName = errors.New("invalid file name")
)

// SanitizeFileName 规范化并校验文件名，防止非法字符和保留名
func SanitizeFileName(name string) (string, error) {
    if name == "" {
        return "", ErrInvalidFileName
    }
    name = filepath.Base(name)
    name = strings.ReplaceAll(name, "\x00", "")
    var cleaned strings.Builder
    for _, r := range name {
        if r >= 32 && r != 127 {
            cleaned.WriteRune(r)
        }
    }
    name = cleaned.String()
    if name == "" || name == "." || name == ".." {
        return "", ErrInvalidFileName
    }
    upperName := strings.ToUpper(name)
    if strings.Contains(upperName, "..") {
        return "", ErrInvalidFileName
    }
    dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
    for _, char := range dangerousChars {
        if strings.Contains(name, char) {
            return "", ErrInvalidFileName
        }
    }
    return name, nil
}

// ValidateZipEntryPath 校验 zip 条目路径是否合法，避免路径遍历等问题
func ValidateZipEntryPath(entryPath string) error {
    cleanPath := filepath.ToSlash(filepath.Clean(entryPath))
    if cleanPath == "." || cleanPath == "" {
        return errors.New("invalid zip entry path: empty path")
    }
    if strings.HasPrefix(cleanPath, "/") {
        return errors.New("invalid zip entry path: absolute path not allowed")
    }
    for _, segment := range strings.Split(cleanPath, "/") {
        if segment == ".." {
            return errors.New("invalid zip entry path: contains path traversal")
        }
    }
    return nil
}
