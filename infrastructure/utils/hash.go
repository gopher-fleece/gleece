package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func Sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
