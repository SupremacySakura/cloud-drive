package repository

import (
	"regexp"
	"testing"
	"time"

	"cloud-drive-backend/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

	cleanup := func() {
		sqlDB.Close()
	}

	return db, mock, cleanup
}

// Test Create File
func TestFileRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	file := &model.FileModel{
		UserID:   1,
		FolderID: 1,
		Name:     "test.txt",
		Size:     1024,
		Type:     "text/plain",
		FileHash: "abc123",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `file_models`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(file)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetFileByHash
func TestFileRepository_GetFileByHash(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `file_models`")).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "folder_id", "name", "size", "type", "file_hash"}).
			AddRow(1, 1, 1, "test.txt", 1024, "text/plain", "abc123"))

	file, err := repo.GetFileByHash("abc123")

	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, "abc123", file.FileHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_GetFileByHash_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `file_models`")).
		WithArgs("nonexistent").
		WillReturnError(gorm.ErrRecordNotFound)

	file, err := repo.GetFileByHash("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test CreateUploadTask
func TestFileRepository_CreateUploadTask(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	task := &model.UploadTask{
		FileName:    "test.txt",
		FileSize:    1024,
		FileHash:    "abc123",
		ChunkSize:   1024,
		TotalChunks: 1,
		UserID:      1,
		FolderID:    1,
		Status:      model.UploadStatusUploading,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `upload_tasks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.CreateUploadTask(task)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUploadTaskByHashAndUserID
func TestFileRepository_GetUploadTaskByHashAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `upload_tasks`")).
		WithArgs("abc123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "file_name", "file_size", "file_hash", "user_id", "status"}).
			AddRow(1, "test.txt", 1024, "abc123", 1, model.UploadStatusUploading))

	task, err := repo.GetUploadTaskByHashAndUserID("abc123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "test.txt", task.FileName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetFileByFileIDAndUserID
func TestFileRepository_GetFileByFileIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `file_models`")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "folder_id", "name", "size", "type", "file_hash"}).
			AddRow(1, 1, 1, "test.txt", 1024, "text/plain", "abc123"))

	file, err := repo.GetFileByFileIDAndUserID(1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, uint(1), file.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test CheckFileExistsInFolder
func TestFileRepository_CheckFileExistsInFolder(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `file_models`")).
		WithArgs("abc123", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.CheckFileExistsInFolder("abc123", 1, 1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_CheckFileExistsInFolder_NotExists(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `file_models`")).
		WithArgs("abc123", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := repo.CheckFileExistsInFolder("abc123", 1, 1)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUploadTaskByIDAndUserID
func TestFileRepository_GetUploadTaskByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `upload_tasks`")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "file_name", "file_size", "user_id"}).
			AddRow(1, "test.txt", 1024, 1))

	task, err := repo.GetUploadTaskByIDAndUserID(1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, uint(1), task.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test UpdateUploadTask
func TestFileRepository_UpdateUploadTask(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	task := &model.UploadTask{
		ID:       1,
		FileName: "test.txt",
		Status:   model.UploadStatusCompleted,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `upload_tasks`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateUploadTask(task)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test MakeDirectory
func TestFileRepository_MakeDirectory(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `folder_models`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := repo.MakeDirectory(0, "newfolder", 1)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetFolderByFolderIDAndUserID
func TestFileRepository_GetFolderByFolderIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `folder_models`")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "name", "user_id"}).
			AddRow(1, 0, "testfolder", 1))

	folder, err := repo.GetFolderByFolderIDAndUserID(1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, folder)
	assert.Equal(t, "testfolder", folder.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test RenameFileByIDAndUserID
func TestFileRepository_RenameFileByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `file_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.RenameFileByIDAndUserID(1, 1, "newname.txt")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_RenameFileByIDAndUserID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `file_models`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.RenameFileByIDAndUserID(999, 1, "newname.txt")

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test DeleteFileByIDAndUserID
func TestFileRepository_DeleteFileByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `file_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteFileByIDAndUserID(1, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_DeleteFileByIDAndUserID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `file_models`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.DeleteFileByIDAndUserID(999, 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetStorageUsedByUserID
func TestFileRepository_GetStorageUsedByUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(size), 0)")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"COALESCE(SUM(size), 0)"}).AddRow(1024000))

	used, err := repo.GetStorageUsedByUserID(1)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1024000), used)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetFileStatsByUserID
func TestFileRepository_GetFileStatsByUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"type", "count", "size"}).
			AddRow("image", 10, 1024000).
			AddRow("document", 5, 512000))

	stats, err := repo.GetFileStatsByUserID(1)

	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "image", stats[0].Type)
	assert.Equal(t, int64(10), stats[0].Count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test CreatePickUpCode
func TestFileRepository_CreatePickUpCode(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	code := &model.PickUpCodeModel{
		Code:        "ABC123",
		UserID:      1,
		Status:      model.PickUpCodeStatusActive,
		MaxDownload: 10,
		ExpireTime:  time.Now().Add(24 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `pick_up_code_models`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := repo.CreatePickUpCode(code)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetPickUpCodeByCode
func TestFileRepository_GetPickUpCodeByCode(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pick_up_code_models`")).
		WithArgs("ABC123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "user_id", "status"}).
			AddRow(1, "ABC123", 1, model.PickUpCodeStatusActive))

	code, err := repo.GetPickUpCodeByCode("ABC123")

	assert.NoError(t, err)
	assert.NotNil(t, code)
	assert.Equal(t, "ABC123", code.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetPickUpCodeListCountByUserID
func TestFileRepository_GetPickUpCodeListCountByUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `pick_up_code_models`")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.GetPickUpCodeListCountByUserID(1)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test DeletePickUpCodeByIDAndUserID
func TestFileRepository_DeletePickUpCodeByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `pick_up_code_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeletePickUpCodeByIDAndUserID(1, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_DeletePickUpCodeByIDAndUserID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `pick_up_code_models`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.DeletePickUpCodeByIDAndUserID(999, 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test CreatePublicShareLink
func TestFileRepository_CreatePublicShareLink(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	link := &model.PublicShareLinkModel{
		Token:  "testtoken123",
		FileID: 1,
		UserID: 1,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `public_share_link_models`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.CreatePublicShareLink(link)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetPublicShareLinkByToken
func TestFileRepository_GetPublicShareLinkByToken(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `public_share_link_models`")).
		WithArgs("testtoken123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "token", "file_id", "user_id"}).
			AddRow(1, "testtoken123", 1, 1))

	link, err := repo.GetPublicShareLinkByToken("testtoken123")

	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, "testtoken123", link.Token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test DeletePublicShareLinkByFileIDAndUserID
func TestFileRepository_DeletePublicShareLinkByFileIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `public_share_link_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeletePublicShareLinkByFileIDAndUserID(1, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetFileByID
func TestFileRepository_GetFileByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `file_models`")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "size", "file_hash"}).
			AddRow(1, 1, "test.txt", 1024, "abc123"))

	file, err := repo.GetFileByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, uint(1), file.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetFolderByID
func TestFileRepository_GetFolderByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `folder_models`")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "name", "user_id"}).
			AddRow(1, 0, "root", 1))

	folder, err := repo.GetFolderByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, folder)
	assert.Equal(t, "root", folder.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test MoveFileByIDAndUserID
func TestFileRepository_MoveFileByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `file_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.MoveFileByIDAndUserID(1, 1, 2)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test RenameFolderByIDAndUserID
func TestFileRepository_RenameFolderByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `folder_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.RenameFolderByIDAndUserID(1, 1, "newfoldername")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test MoveFolderByIDAndUserID
func TestFileRepository_MoveFolderByIDAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `folder_models`")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.MoveFolderByIDAndUserID(1, 1, 2)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetDirectChildFoldersByParentAndUserID
func TestFileRepository_GetDirectChildFoldersByParentAndUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `folder_models`")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "name", "user_id"}).
			AddRow(2, 1, "subfolder1", 1).
			AddRow(3, 1, "subfolder2", 1))

	folders, err := repo.GetDirectChildFoldersByParentAndUserID(1, 1)

	assert.NoError(t, err)
	assert.Len(t, folders, 2)
	assert.Equal(t, "subfolder1", folders[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
