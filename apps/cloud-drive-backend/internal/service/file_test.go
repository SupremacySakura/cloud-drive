package service

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/repository"
)

func TestSanitizeFileName_ValidName(t *testing.T) {
	name := "正常的文件.txt"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "正常的文件.txt", result)
}

func TestSanitizeFileName_EmptyName(t *testing.T) {
	result, err := sanitizeFileName("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileName, err)
	assert.Empty(t, result)
}

func TestSanitizeFileName_PathTraversal(t *testing.T) {
	name := "../../../etc/passwd"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "passwd", result)
}

func TestSanitizeFileName_DangerousChars(t *testing.T) {
	dangerous := []string{
		"file|name",
		"file*name",
		"file?name",
		"file<name",
		"file>name",
		"file|name",
	}
	for _, name := range dangerous {
		result, err := sanitizeFileName(name)
		assert.Error(t, err, "Expected error for: %s", name)
		assert.Equal(t, ErrInvalidFileName, err)
		assert.Empty(t, result)
	}
}

func TestSanitizeFileName_ReservedName(t *testing.T) {
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "LPT1"}
	for _, name := range reserved {
		result, err := sanitizeFileName(name)
		assert.Error(t, err, "Expected error for reserved name: %s", name)
		assert.Equal(t, ErrInvalidFileName, err)
		assert.Empty(t, result)
	}
}

func TestSanitizeFileName_DotNames(t *testing.T) {
	dotNames := []string{".", ".."}
	for _, name := range dotNames {
		result, err := sanitizeFileName(name)
		assert.Error(t, err, "Expected error for dot name: %s", name)
		assert.Equal(t, ErrInvalidFileName, err)
		assert.Empty(t, result)
	}
}

func TestSanitizeFileName_WithExtension(t *testing.T) {
	name := "document.pdf"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "document.pdf", result)
}

func TestSanitizeFileName_WithDangerousExt(t *testing.T) {
	name := "file.PRN"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "file.PRN", result)
}

// Test IsAllowedMIMEType
func TestIsAllowedMIMEType_ValidTypes(t *testing.T) {
	svc := &fileService{}

	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"video/mp4",
		"video/avi",
		"text/plain",
		"text/html",
		"application/pdf",
		"application/zip",
	}

	for _, mimeType := range validTypes {
		result := svc.IsAllowedMIMEType(mimeType)
		assert.True(t, result, "Expected %s to be allowed", mimeType)
	}
}

func TestIsAllowedMIMEType_InvalidTypes(t *testing.T) {
	svc := &fileService{}

	invalidTypes := []string{
		"application/x-executable",
		"application/x-msdownload",
		"text/javascript",
		"application/javascript",
		"application/x-shockwave-flash",
		"",
	}

	for _, mimeType := range invalidTypes {
		result := svc.IsAllowedMIMEType(mimeType)
		assert.False(t, result, "Expected %s to be blocked", mimeType)
	}
}

// Test sanitizeStorageFileExt
func TestSanitizeStorageFileExt_ValidExt(t *testing.T) {
	ext := sanitizeStorageFileExt("document.pdf")
	assert.Equal(t, ".pdf", ext)
}

func TestSanitizeStorageFileExt_NoExt(t *testing.T) {
	ext := sanitizeStorageFileExt("document")
	assert.Equal(t, "", ext)
}

func TestSanitizeStorageFileExt_EmptyName(t *testing.T) {
	ext := sanitizeStorageFileExt("")
	assert.Equal(t, "", ext)
}

func TestSanitizeStorageFileExt_DangerousChars(t *testing.T) {
	ext := sanitizeStorageFileExt("file<script>.txt")
	assert.NotContains(t, ext, "<")
	assert.NotContains(t, ext, ">")
}

// Test validateZipEntryPath
func TestValidateZipEntryPath_ValidPath(t *testing.T) {
	paths := []string{
		"folder/file.txt",
		"file.txt",
		"folder/subfolder/file.txt",
	}

	for _, path := range paths {
		err := validateZipEntryPath(path)
		assert.NoError(t, err, "Path %s should be valid", path)
	}
}

func TestValidateZipEntryPath_PathTraversal(t *testing.T) {
	paths := []string{
		"../file.txt",
		"folder/../../file.txt",
		"folder/../file.txt",
	}

	for _, path := range paths {
		err := validateZipEntryPath(path)
		assert.Error(t, err, "Path %s should be invalid", path)
	}
}

func TestValidateZipEntryPath_AbsolutePath(t *testing.T) {
	err := validateZipEntryPath("/absolute/path/file.txt")
	assert.Error(t, err)
}

func TestValidateZipEntryPath_EmptyPath(t *testing.T) {
	err := validateZipEntryPath("")
	assert.Error(t, err)
}

// Test fileService errors
func TestFileService_Errors(t *testing.T) {
	// Verify error variables exist and have correct messages
	assert.NotNil(t, ErrPickupCodeExpired)
	assert.NotNil(t, ErrPickupTargetNotFound)
	assert.NotNil(t, ErrPickupEmptyFolder)
	assert.NotNil(t, ErrPublicShareNotFound)
	assert.NotNil(t, ErrInvalidFileName)
	assert.NotNil(t, ErrStorageQuotaExceeded)
	assert.NotNil(t, ErrChunkSizeMismatch)
	assert.NotNil(t, ErrInvalidMIMEType)
}

// Test allowedMIMETypes map
func TestAllowedMIMETypes(t *testing.T) {
	// Verify the map is properly initialized
	assert.True(t, allowedMIMETypes["image/"])
	assert.True(t, allowedMIMETypes["video/"])
	assert.True(t, allowedMIMETypes["application/pdf"])
	assert.True(t, allowedMIMETypes["application/zip"])
	assert.True(t, allowedMIMETypes["text/"])
}

// Test reservedNames map
func TestReservedNames(t *testing.T) {
	// Verify reserved Windows names are blocked
	assert.True(t, reservedNames["CON"])
	assert.True(t, reservedNames["PRN"])
	assert.True(t, reservedNames["AUX"])
	assert.True(t, reservedNames["NUL"])
	assert.True(t, reservedNames["COM1"])
	assert.True(t, reservedNames["LPT1"])
}

// Test sanitizeFileName with null bytes
func TestSanitizeFileName_NullBytes(t *testing.T) {
	name := "file\x00name.txt"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.NotContains(t, result, "\x00")
	assert.Equal(t, "filename.txt", result)
}

// Test sanitizeFileName with control characters
func TestSanitizeFileName_ControlChars(t *testing.T) {
	name := "file\x01\x02\x03name.txt"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "filename.txt", result)
}

// Test sanitizeFileName with spaces
func TestSanitizeFileName_WithSpaces(t *testing.T) {
	name := "file name with spaces.txt"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "file name with spaces.txt", result)
}

// Test sanitizeFileName with unicode
func TestSanitizeFileName_Unicode(t *testing.T) {
	names := []string{
		"文件.txt",
		"文件_日本語.doc",
		"файл.pdf",
		"😀emoji.txt",
	}

	for _, name := range names {
		result, err := sanitizeFileName(name)
		assert.NoError(t, err, "Name %s should be valid", name)
		assert.NotEmpty(t, result)
	}
}

// Test sanitizeFileName very long name
func TestSanitizeFileName_VeryLongName(t *testing.T) {
	// Create a very long name
	longName := ""
	for i := 0; i < 100; i++ {
		longName += "a"
	}
	longName += ".txt"

	result, err := sanitizeFileName(longName)
	assert.NoError(t, err)
	assert.Equal(t, longName, result)
}

// Test sanitizeFileName multiple extensions
func TestSanitizeFileName_MultipleExtensions(t *testing.T) {
	name := "file.tar.gz"
	result, err := sanitizeFileName(name)
	assert.NoError(t, err)
	assert.Equal(t, "file.tar.gz", result)
}

// Test sanitizeFileName case sensitivity
func TestSanitizeFileName_CaseSensitivity(t *testing.T) {
	// Reserved names should be case-insensitive
	_, err := sanitizeFileName("con")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileName, err)

	_, err = sanitizeFileName("CON")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileName, err)

	_, err = sanitizeFileName("Con")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileName, err)
}

func TestBuildFileAbsolutePath_UsesCanonicalHashPath(t *testing.T) {
	tempDir := t.TempDir()
	fileHash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	storageDir := filepath.Join(tempDir, fileHash[0:2], fileHash[2:4])
	assert.NoError(t, os.MkdirAll(storageDir, 0755))

	canonicalPath := filepath.Join(storageDir, fileHash)
	assert.NoError(t, os.WriteFile(canonicalPath, []byte("content"), 0644))

	svc := &fileService{FileServiceOptions: FileServiceOptions{FileStoragePath: tempDir}}
	path, err := svc.BuildFileAbsolutePath(&model.FileModel{
		FileHash: fileHash,
		Name:     "renamed.pdf",
	})

	assert.NoError(t, err)
	assert.Equal(t, canonicalPath, path)
}

func TestBuildFileAbsolutePath_FallsBackToLegacyExtensionPath(t *testing.T) {
	tempDir := t.TempDir()
	fileHash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	storageDir := filepath.Join(tempDir, fileHash[0:2], fileHash[2:4])
	assert.NoError(t, os.MkdirAll(storageDir, 0755))

	legacyPath := filepath.Join(storageDir, fileHash+".txt")
	assert.NoError(t, os.WriteFile(legacyPath, []byte("content"), 0644))

	svc := &fileService{FileServiceOptions: FileServiceOptions{FileStoragePath: tempDir}}
	path, err := svc.BuildFileAbsolutePath(&model.FileModel{
		FileHash: fileHash,
		Name:     "renamed.pdf",
	})

	assert.NoError(t, err)
	assert.Equal(t, legacyPath, path)
}

func setupServiceMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		DriverName:                "mysql",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return db, mock, func() {
		_ = sqlDB.Close()
	}
}

func TestDownloadByPickUpCode_ReservesDownloadBeforeStreaming(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()

	tempDir := t.TempDir()
	fileHash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	storageDir := filepath.Join(tempDir, fileHash[0:2], fileHash[2:4])
	assert.NoError(t, os.MkdirAll(storageDir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(storageDir, fileHash), []byte("secret"), 0644))

	expireTime := time.Now().Add(time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pick_up_code_models`")).
		WithArgs("ABC123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "file_id", "folder_id", "download", "max_download", "expire_time", "code", "status", "user_id"}).
			AddRow(9, model.PickUpTargetTypeFile, uint(7), nil, 0, 1, expireTime, "ABC123", model.PickUpCodeStatusActive, 42))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `file_models`")).
		WithArgs(uint(7), uint(42), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "file_hash", "type", "size", "folder_id", "user_id"}).
			AddRow(7, "secret.txt", fileHash, "document", 6, 0, 42))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pick_up_code_models`")).
		WithArgs(uint(9), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "file_id", "folder_id", "download", "max_download", "expire_time", "code", "status", "user_id"}).
			AddRow(9, model.PickUpTargetTypeFile, uint(7), nil, 1, 1, expireTime, "ABC123", model.PickUpCodeStatusActive, 42))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `pick_up_code_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	svc := &fileService{
		FileRepository: repository.NewFileRepository(db),
		FileServiceOptions: FileServiceOptions{
			FileStoragePath: tempDir,
		},
	}
	var writer bytes.Buffer
	err := svc.DownloadByPickUpCode("ABC123", &writer, nil)

	assert.ErrorIs(t, err, ErrPickupCodeExpired)
	assert.Empty(t, writer.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}
