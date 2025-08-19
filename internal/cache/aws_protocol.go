package cache

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	lib "github.com/thelonelyghost/vault-aws-credential-protocol/internal/aws"
)

var (
	cacheDir string
)

func init() {
	cacheDirBase, _ := os.UserCacheDir()
	cacheDir = filepath.Join(cacheDirBase, "vault-aws-credential-protocol")
}

func ensureCacheDir() error {
	return os.MkdirAll(cacheDir, 0700)
}

func generateFilename(key string) string {
	keyHash := sha256.Sum256([]byte(key))
	filename := base32.StdEncoding.EncodeToString(keyHash[:])

	return filepath.Join(cacheDir, filename)
}

func encodePayload(payload string) string {
	return base64.URLEncoding.EncodeToString([]byte(payload))
}
func decodePayload(payload string) (string, error) {
	output, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func SetCache(key string, data lib.AwsCredentialProtocolOutput) error {
	err := ensureCacheDir()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(
		generateFilename(key),
		[]byte(encodePayload(string(payload))),
		0600,
	)
}

func GetCache(key string) (lib.AwsCredentialProtocolOutput, error) {
	data := lib.AwsCredentialProtocolOutput{}
	contents, err := os.ReadFile(generateFilename(key))
	if err != nil {
		return data, err
	}

	payload, err := decodePayload(string(contents))
	if err != nil {
		return data, err
	}

	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return data, err
	}

	now := time.Now()
	expiry, err := time.Parse(time.RFC3339, data.Expiration)

	if err != nil {
		return data, err
	}

	// If less than 10 minutes left before expiring, treat it as needing
	// refreshed and do not acknowledge the cached credentials
	minTimeLeftForCacheHit := 10 * time.Minute

	if !now.Before(expiry) {
		return data, ErrExpiredCredentials
	}

	if !now.Add(minTimeLeftForCacheHit).Before(expiry) {
		return data, ErrNearExpiration
	}

	return data, nil
}

func ClearCache() error {
	return os.RemoveAll(cacheDir)
}
