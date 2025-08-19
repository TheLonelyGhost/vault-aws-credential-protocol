package cache

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	lib "github.com/thelonelyghost/vault-aws-credential-protocol/internal/aws"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomKey() string {
	b := make([]rune, 50)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func TestCacheDirectoryExists(t *testing.T) {
	base := cacheDir

	if _, err := os.Stat(base); err == nil {
		os.RemoveAll(base)
	}

	if _, err := os.Stat(base); err == nil {
		t.Fatal("Cache directory still exists after attempting to remove it")
	}
	ensureCacheDir()
	if _, err := os.Stat(base); err != nil {
		t.Fatal(err)
	}
}

func TestUniqueCacheKey(t *testing.T) {
	first := generateFilename(randomKey())
	second := generateFilename(randomKey())

	if first == second {
		t.Fatal("Unable to generate unique cache keys for different input")
	}

	if filepath.Dir(first) != cacheDir {
		t.Fatal("Generated cache key is not part of the allowed cache directory")
	}
}

func TestSerialization(t *testing.T) {
	// doesn't matter what this is, it just can't be a real ARN so
	// that we don't mess with the existing cache
	key := randomKey()
	obj := lib.AwsCredentialProtocolOutput{}

	defer os.Remove(generateFilename(key))
	err := SetCache(key, obj)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.ReadFile(generateFilename(key)); err != nil {
		t.Fatal(err)
	}
}

func TestDeserialization(t *testing.T) {
	// doesn't matter what this is, it just can't be a real ARN so
	// that we don't mess with the existing cache
	key := randomKey()
	obj := lib.AwsCredentialProtocolOutput{
		AccessKeyId: randomKey(),
		Expiration:  time.Now().AddDate(0, 0, 2).Format(time.RFC3339),
	}

	defer os.Remove(generateFilename(key))
	err := SetCache(key, obj)
	if err != nil {
		t.Fatal(err)
	}

	otherObj, err := GetCache(key)
	if err != nil {
		t.Fatal(err)
	}

	if obj.AccessKeyId != otherObj.AccessKeyId {
		t.Fatal("Deserialized test data is not the same value as when it was serialized")
	}
}

func TestAlmostExpiredCache(t *testing.T) {
	key := randomKey()
	obj := lib.AwsCredentialProtocolOutput{
		AccessKeyId: randomKey(),
		Expiration:  time.Now().Add(5 * time.Minute).Format(time.RFC3339),
	}

	defer os.Remove(generateFilename(key))
	err := SetCache(key, obj)
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetCache(key)
	if err == nil {
		t.Fatal("Failed to skip almost-expired credentials")
	}

	if !errors.Is(err, ErrNearExpiration) {
		t.Fatalf("Error served is not the one expected: %s", err)
	}
}

func TestExpiredCache(t *testing.T) {
	key := randomKey()
	obj := lib.AwsCredentialProtocolOutput{
		AccessKeyId: randomKey(),
		Expiration:  time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
	}

	defer os.Remove(generateFilename(key))
	err := SetCache(key, obj)
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetCache(key)
	if err == nil {
		t.Fatal("Failed to skip expired credentials")
	}

	if !errors.Is(err, ErrExpiredCredentials) {
		t.Fatalf("Error served is not the one expected: %s", err)
	}
}
