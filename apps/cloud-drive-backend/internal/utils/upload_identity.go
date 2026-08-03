package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"cloud-drive-backend/internal/model"
)

var ErrInvalidUploadFileName = errors.New("invalid upload file name")

var reservedUploadFileNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true,
	"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
	"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

func NormalizeUploadFileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrInvalidUploadFileName
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
		return "", ErrInvalidUploadFileName
	}
	upperName := strings.ToUpper(name)
	if reservedUploadFileNames[upperName] {
		return "", ErrInvalidUploadFileName
	}
	if ext := filepath.Ext(upperName); ext != "" {
		baseName := strings.TrimSuffix(upperName, ext)
		if reservedUploadFileNames[baseName] {
			return "", ErrInvalidUploadFileName
		}
	}
	for _, char := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		if strings.Contains(name, char) {
			return "", ErrInvalidUploadFileName
		}
	}
	return name, nil
}

func BuildUploadIdentity(task *model.UploadTask) (idempotencyKey, requestHash, normalizedName string, err error) {
	normalizedName, err = NormalizeUploadFileName(task.FileName)
	if err != nil {
		return "", "", "", err
	}

	identityPayload, err := json.Marshal(struct {
		UserID   uint   `json:"user_id"`
		FolderID uint   `json:"folder_id"`
		FileName string `json:"file_name"`
		FileHash string `json:"file_hash"`
	}{task.UserID, task.FolderID, normalizedName, strings.ToLower(task.FileHash)})
	if err != nil {
		return "", "", "", err
	}
	identitySum := sha256.Sum256(identityPayload)

	requestPayload, err := json.Marshal(struct {
		FileSize    uint64 `json:"file_size"`
		ChunkSize   int    `json:"chunk_size"`
		TotalChunks int    `json:"total_chunks"`
		FileType    string `json:"file_type"`
	}{task.FileSize, task.ChunkSize, task.TotalChunks, task.FileType})
	if err != nil {
		return "", "", "", err
	}
	requestSum := sha256.Sum256(requestPayload)

	return hex.EncodeToString(identitySum[:]), hex.EncodeToString(requestSum[:]), normalizedName, nil
}
