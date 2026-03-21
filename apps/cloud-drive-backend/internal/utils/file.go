package utils

import (
	"cloud-drive-backend/internal/model"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func HasAllChunks(uploaded model.IntSlice, totalChunks int) bool {
	if totalChunks <= 0 {
		return false
	}
	seen := make([]bool, totalChunks)
	for _, idx := range uploaded {
		if idx < 0 || idx >= totalChunks {
			return false
		}
		seen[idx] = true
	}
	for _, ok := range seen {
		if !ok {
			return false
		}
	}
	return true
}

func VerifyFileSHA256(filePath, expectHash string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == expectHash, nil
}
