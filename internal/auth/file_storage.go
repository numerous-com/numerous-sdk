package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"numerous.com/cli/internal/output"
)

const (
	numerousDir      = ".numerous"
	tokenFile        = ".token"
	gitignoreFile    = ".gitignore"
	filePermissions  = 0o600 // Owner read/write only
	dirPermissions   = 0o700 // Owner read/write/execute only
	gitignoreContent = "# Numerous CLI token files\n.token\n"
)

var (
	ErrDirectoryNotWritable = errors.New("directory is not writable")
	ErrFileCorrupted        = errors.New("token file is corrupted")
	ErrFileNotFound         = errors.New("token file not found")
	ErrInvalidPermissions   = errors.New("unable to set secure file permissions")
	ErrUserDeclinedConsent  = errors.New("user declined file storage consent")
)

// FileStorage handles secure file-based token storage
type FileStorage struct {
	baseDir string
}

// TokenData represents the structure of stored token data
type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Tenant       string `json:"tenant"`
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage() *FileStorage {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not available
		homeDir = "."
	}

	return &FileStorage{
		baseDir: filepath.Join(homeDir, numerousDir),
	}
}

// CreateTokenDirectory creates the .numerous directory with proper permissions
func (fs *FileStorage) CreateTokenDirectory() error {
	// Check if directory exists
	if _, err := os.Stat(fs.baseDir); os.IsNotExist(err) {
		// Create directory with secure permissions
		if err := os.MkdirAll(fs.baseDir, dirPermissions); err != nil {
			return fmt.Errorf("%w: %v", ErrDirectoryNotWritable, err)
		}
	}

	// Verify directory is writable
	if !fs.isDirectoryWritable() {
		return ErrDirectoryNotWritable
	}

	// Create .gitignore file to prevent accidental commits
	if err := fs.createGitignoreFile(); err != nil {
		// Non-fatal error - log but don't fail
		output.PrintErrorDetails("Warning: Could not create .gitignore file", err)
	}

	return nil
}

// StoreToken stores the token data in a secure file
func (fs *FileStorage) StoreToken(accessToken, refreshToken, tenant string) error {
	if err := fs.CreateTokenDirectory(); err != nil {
		return err
	}

	tokenData := TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Tenant:       tenant,
	}

	data, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("failed to serialize token data: %v", err)
	}

	tokenPath := filepath.Join(fs.baseDir, tokenFile)

	// Write file with secure permissions
	if err := os.WriteFile(tokenPath, data, filePermissions); err != nil {
		return fmt.Errorf("failed to write token file: %v", err)
	}

	// Verify file permissions were set correctly
	if err := fs.verifyFilePermissions(tokenPath); err != nil {
		// Try to remove the file if permissions couldn't be set properly
		os.Remove(tokenPath)
		return err
	}

	return nil
}

// RetrieveToken reads token data from the file
func (fs *FileStorage) RetrieveToken() (*TokenData, error) {
	tokenPath := filepath.Join(fs.baseDir, tokenFile)

	// Check if file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}

	// Read file content
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %v", err)
	}

	// Parse JSON data
	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileCorrupted, err)
	}

	// Validate required fields
	if tokenData.AccessToken == "" || tokenData.RefreshToken == "" || tokenData.Tenant == "" {
		return nil, ErrFileCorrupted
	}

	return &tokenData, nil
}

// DeleteToken removes the token file
func (fs *FileStorage) DeleteToken() error {
	tokenPath := filepath.Join(fs.baseDir, tokenFile)

	// Check if file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		// File doesn't exist, which is fine for deletion
		return nil
	}

	// Remove the file
	if err := os.Remove(tokenPath); err != nil {
		return fmt.Errorf("failed to delete token file: %v", err)
	}

	return nil
}

// GetTokenDirectory returns the directory where tokens are stored
func (fs *FileStorage) GetTokenDirectory() string {
	return fs.baseDir
}

// GetTokenFilePath returns the full path to the token file
func (fs *FileStorage) GetTokenFilePath() string {
	return filepath.Join(fs.baseDir, tokenFile)
}

// isDirectoryWritable checks if the directory is writable
func (fs *FileStorage) isDirectoryWritable() bool {
	testFile := filepath.Join(fs.baseDir, ".test_write")

	// Try to create a test file
	if err := os.WriteFile(testFile, []byte("test"), filePermissions); err != nil {
		return false
	}

	// Clean up test file
	os.Remove(testFile)

	return true
}

// verifyFilePermissions checks if file permissions are set correctly
func (fs *FileStorage) verifyFilePermissions(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Check if permissions are 0600 (owner read/write only)
	if info.Mode().Perm() != filePermissions {
		return ErrInvalidPermissions
	}

	return nil
}

// createGitignoreFile creates a .gitignore file to prevent token commits
func (fs *FileStorage) createGitignoreFile() error {
	gitignorePath := filepath.Join(fs.baseDir, gitignoreFile)

	// Check if .gitignore already exists
	if _, err := os.Stat(gitignorePath); err == nil {
		return nil // File already exists
	}

	// Create .gitignore file
	return os.WriteFile(gitignorePath, []byte(gitignoreContent), filePermissions)
}
