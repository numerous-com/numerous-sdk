package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	fs := NewFileStorage()

	if fs == nil {
		t.Fatal("NewFileStorage() returned nil")
	}

	// Should contain .numerous in the path
	if !filepath.IsAbs(fs.baseDir) && fs.baseDir != ".numerous" {
		t.Errorf("Expected baseDir to be absolute or '.numerous', got: %s", fs.baseDir)
	}
}

func TestCreateTokenDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Test creating directory
	err = fs.CreateTokenDirectory()
	if err != nil {
		t.Errorf("CreateTokenDirectory() failed: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(fs.baseDir)
	if err != nil {
		t.Errorf("Directory was not created: %v", err)
	}

	// Verify directory permissions
	if info.Mode().Perm() != dirPermissions {
		t.Errorf("Directory permissions incorrect. Expected: %o, Got: %o",
			dirPermissions, info.Mode().Perm())
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(fs.baseDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore file was not created")
	}
}

func TestStoreToken(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Test data
	accessToken := "test-access-token"
	refreshToken := "test-refresh-token"
	tenant := "test-tenant"

	// Store token
	err = fs.StoreToken(accessToken, refreshToken, tenant)
	if err != nil {
		t.Errorf("StoreToken() failed: %v", err)
	}

	// Verify file exists
	tokenPath := fs.GetTokenFilePath()
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Errorf("Token file was not created: %v", err)
	}

	// Verify file permissions
	if info.Mode().Perm() != filePermissions {
		t.Errorf("File permissions incorrect. Expected: %o, Got: %o",
			filePermissions, info.Mode().Perm())
	}

	// Verify file content
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Errorf("Failed to read token file: %v", err)
	}

	var tokenData TokenData
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		t.Errorf("Failed to parse token file: %v", err)
	}

	if tokenData.AccessToken != accessToken {
		t.Errorf("Access token mismatch. Expected: %s, Got: %s", accessToken, tokenData.AccessToken)
	}
	if tokenData.RefreshToken != refreshToken {
		t.Errorf("Refresh token mismatch. Expected: %s, Got: %s", refreshToken, tokenData.RefreshToken)
	}
	if tokenData.Tenant != tenant {
		t.Errorf("Tenant mismatch. Expected: %s, Got: %s", tenant, tokenData.Tenant)
	}
}

func TestRetrieveToken(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Test data
	accessToken := "test-access-token"
	refreshToken := "test-refresh-token"
	tenant := "test-tenant"

	// First store token
	err = fs.StoreToken(accessToken, refreshToken, tenant)
	if err != nil {
		t.Fatalf("StoreToken() failed: %v", err)
	}

	// Then retrieve it
	tokenData, err := fs.RetrieveToken()
	if err != nil {
		t.Errorf("RetrieveToken() failed: %v", err)
	}

	if tokenData == nil {
		t.Fatal("RetrieveToken() returned nil token data")
	}

	if tokenData.AccessToken != accessToken {
		t.Errorf("Access token mismatch. Expected: %s, Got: %s", accessToken, tokenData.AccessToken)
	}
	if tokenData.RefreshToken != refreshToken {
		t.Errorf("Refresh token mismatch. Expected: %s, Got: %s", refreshToken, tokenData.RefreshToken)
	}
	if tokenData.Tenant != tenant {
		t.Errorf("Tenant mismatch. Expected: %s, Got: %s", tenant, tokenData.Tenant)
	}
}

func TestRetrieveTokenFileNotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Try to retrieve token when file doesn't exist
	tokenData, err := fs.RetrieveToken()
	if err != ErrFileNotFound {
		t.Errorf("Expected ErrFileNotFound, got: %v", err)
	}
	if tokenData != nil {
		t.Error("Expected nil token data when file not found")
	}
}

func TestRetrieveTokenCorruptedFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Create directory first
	err = fs.CreateTokenDirectory()
	if err != nil {
		t.Fatalf("CreateTokenDirectory() failed: %v", err)
	}

	// Create corrupted token file
	tokenPath := fs.GetTokenFilePath()
	err = os.WriteFile(tokenPath, []byte("invalid json"), filePermissions)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Try to retrieve token
	tokenData, err := fs.RetrieveToken()
	if !errors.Is(err, ErrFileCorrupted) {
		t.Errorf("Expected ErrFileCorrupted, got: %v", err)
	}
	if tokenData != nil {
		t.Error("Expected nil token data when file is corrupted")
	}
}

func TestRetrieveTokenIncompleteData(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Create directory first
	err = fs.CreateTokenDirectory()
	if err != nil {
		t.Fatalf("CreateTokenDirectory() failed: %v", err)
	}

	// Create token file with incomplete data
	incompleteData := TokenData{
		AccessToken: "test-token",
		// Missing RefreshToken and Tenant
	}
	data, _ := json.Marshal(incompleteData)

	tokenPath := fs.GetTokenFilePath()
	err = os.WriteFile(tokenPath, data, filePermissions)
	if err != nil {
		t.Fatalf("Failed to create incomplete file: %v", err)
	}

	// Try to retrieve token
	tokenData, err := fs.RetrieveToken()
	if err != ErrFileCorrupted {
		t.Errorf("Expected ErrFileCorrupted for incomplete data, got: %v", err)
	}
	if tokenData != nil {
		t.Error("Expected nil token data when file has incomplete data")
	}
}

func TestDeleteToken(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// First store token
	err = fs.StoreToken("access", "refresh", "tenant")
	if err != nil {
		t.Fatalf("StoreToken() failed: %v", err)
	}

	// Verify file exists
	tokenPath := fs.GetTokenFilePath()
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file should exist before deletion")
	}

	// Delete token
	err = fs.DeleteToken()
	if err != nil {
		t.Errorf("DeleteToken() failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("Token file should not exist after deletion")
	}
}

func TestDeleteTokenFileNotExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: filepath.Join(tempDir, ".numerous"),
	}

	// Try to delete non-existent token file
	err = fs.DeleteToken()
	if err != nil {
		t.Errorf("DeleteToken() should not fail when file doesn't exist: %v", err)
	}
}

func TestGetTokenDirectory(t *testing.T) {
	tempDir := "/tmp/test"
	fs := &FileStorage{
		baseDir: tempDir,
	}

	if fs.GetTokenDirectory() != tempDir {
		t.Errorf("GetTokenDirectory() returned wrong path. Expected: %s, Got: %s",
			tempDir, fs.GetTokenDirectory())
	}
}

func TestGetTokenFilePath(t *testing.T) {
	tempDir := "/tmp/test"
	fs := &FileStorage{
		baseDir: tempDir,
	}

	expected := filepath.Join(tempDir, tokenFile)
	if fs.GetTokenFilePath() != expected {
		t.Errorf("GetTokenFilePath() returned wrong path. Expected: %s, Got: %s",
			expected, fs.GetTokenFilePath())
	}
}

func TestIsDirectoryWritable(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: tempDir,
	}

	// Should be writable
	if !fs.isDirectoryWritable() {
		t.Error("Temporary directory should be writable")
	}
}

func TestVerifyFilePermissions(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "numerous-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := &FileStorage{
		baseDir: tempDir,
	}

	// Create a file with correct permissions
	testFile := filepath.Join(tempDir, "test")
	err = os.WriteFile(testFile, []byte("test"), filePermissions)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify permissions
	err = fs.verifyFilePermissions(testFile)
	if err != nil {
		t.Errorf("verifyFilePermissions() failed for correct permissions: %v", err)
	}

	// Create a file with wrong permissions
	wrongFile := filepath.Join(tempDir, "wrong")
	err = os.WriteFile(wrongFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file with wrong permissions: %v", err)
	}

	// Verify it fails
	err = fs.verifyFilePermissions(wrongFile)
	if err != ErrInvalidPermissions {
		t.Errorf("verifyFilePermissions() should fail for wrong permissions, got: %v", err)
	}
}
