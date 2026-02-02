// pkg/cache/cache.go
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rohansx/vgx/pkg/types"
)

var (
	cacheDir string
	cacheMu  sync.RWMutex
)

// CacheEntry represents a cached scan result
type CacheEntry struct {
	FilePath       string               `json:"file_path"`
	FileHash       string               `json:"file_hash"`
	Vulnerabilities []types.Vulnerability `json:"vulnerabilities"`
	Timestamp      time.Time            `json:"timestamp"`
}

// Initialize sets up the cache directory
func Initialize(dir string) error {
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "vgx-cache")
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheDir = dir
	return nil
}

// Get retrieves a cached scan result if available
func Get(filePath string) ([]types.Vulnerability, bool, error) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	if cacheDir == "" {
		return nil, false, fmt.Errorf("cache not initialized")
	}

	// Calculate file hash
	hash, err := hashFile(filePath)
	if err != nil {
		return nil, false, err
	}

	// Create cache key from file path
	cacheKey := createCacheKey(filePath)
	cacheFilePath := filepath.Join(cacheDir, cacheKey+".json")

	// Check if cache file exists
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		return nil, false, nil
	}

	// Read and parse cache file
	data, err := ioutil.ReadFile(cacheFilePath)
	if err != nil {
		return nil, false, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, err
	}

	// Check if file has changed
	if entry.FileHash != hash {
		return nil, false, nil
	}

	// Check if cache is expired (optional, 24 hour TTL)
	if time.Since(entry.Timestamp) > 24*time.Hour {
		return nil, false, nil
	}

	return entry.Vulnerabilities, true, nil
}

// Store saves scan results to cache
func Store(filePath string, vulnerabilities []types.Vulnerability) error {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cacheDir == "" {
		return fmt.Errorf("cache not initialized")
	}

	// Calculate file hash
	hash, err := hashFile(filePath)
	if err != nil {
		return err
	}

	// Create cache entry
	entry := CacheEntry{
		FilePath:       filePath,
		FileHash:       hash,
		Vulnerabilities: vulnerabilities,
		Timestamp:      time.Now(),
	}

	// Serialize to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Create cache key from file path
	cacheKey := createCacheKey(filePath)
	cacheFilePath := filepath.Join(cacheDir, cacheKey+".json")

	// Write to cache file
	return ioutil.WriteFile(cacheFilePath, data, 0644)
}

// Clear removes all cache entries
func Clear() error {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cacheDir == "" {
		return fmt.Errorf("cache not initialized")
	}

	// Read all files in cache dir
	files, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	// Delete each cache file
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			if err := os.Remove(filepath.Join(cacheDir, file.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

// Invalidate removes a specific cache entry
func Invalidate(filePath string) error {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cacheDir == "" {
		return fmt.Errorf("cache not initialized")
	}

	cacheKey := createCacheKey(filePath)
	cacheFilePath := filepath.Join(cacheDir, cacheKey+".json")

	// Delete if exists
	if _, err := os.Stat(cacheFilePath); err == nil {
		return os.Remove(cacheFilePath)
	}

	return nil
}

// hashFile calculates SHA-256 hash of a file
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createCacheKey generates a safe filename from a file path
func createCacheKey(filePath string) string {
	hash := sha256.Sum256([]byte(filePath))
	return hex.EncodeToString(hash[:])
}
