package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

var ErrModelKeyUnavailable = errors.New("模型密钥加密未配置")

type ModelCredentials struct{ aead cipher.AEAD }

func NewModelCredentials(encoded string) (*ModelCredentials, error) {
	if encoded == "" {
		return &ModelCredentials{}, nil
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, errors.New("MODEL_CONFIG_ENCRYPTION_KEY 必须是 32 字节随机密钥的 Base64 编码")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	return &ModelCredentials{aead: aead}, err
}

func credentialContext(user uint, provider string) []byte {
	return []byte(fmt.Sprintf("zhixu:model-key:v1:%d:%s", user, provider))
}

func (c *ModelCredentials) Encrypt(user uint, provider, key string) (string, error) {
	if c == nil || c.aead == nil {
		return "", ErrModelKeyUnavailable
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(key), credentialContext(user, provider))
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt is reserved for future authenticated internal model calls, never public responses.
func (c *ModelCredentials) Decrypt(user uint, provider, value string) (string, error) {
	if c == nil || c.aead == nil {
		return "", ErrModelKeyUnavailable
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(data) < c.aead.NonceSize()+c.aead.Overhead() {
		return "", errors.New("模型凭据无法解密")
	}
	size := c.aead.NonceSize()
	plain, err := c.aead.Open(nil, data[:size], data[size:], credentialContext(user, provider))
	if err != nil {
		return "", errors.New("模型凭据无法解密")
	}
	return string(plain), nil
}
