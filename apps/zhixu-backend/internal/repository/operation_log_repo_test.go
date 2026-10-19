package repository

import (
	"regexp"
	"testing"
	"time"

	"zhixu-backend/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestOperationLogRepository_CreateOperationLog(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	log := &model.OperationLogModel{
		UserID:     1,
		Action:     model.OperationLogActionUpload,
		TargetType: model.OperationLogTargetTypeFile,
		TargetID:   10,
		TargetName: "report.pdf",
		FolderName: "工作",
		Detail:     `{"size":1024}`,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `operation_log_models`")).
		WillReturnResult(sqlmock.NewResult(100, 1))
	mock.ExpectCommit()

	err := repo.CreateOperationLog(log)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_GetOperationLogsByUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "action", "target_type", "target_id", "target_name", "folder_name", "detail", "created_at"}).
		AddRow(30, 1, model.OperationLogActionDelete, model.OperationLogTargetTypeFile, 9, "old.txt", "", "", now).
		AddRow(25, 1, model.OperationLogActionUpload, model.OperationLogTargetTypeFile, 10, "report.pdf", "工作", `{"size":1024}`, now.Add(-time.Hour))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `operation_log_models`")).
		WithArgs(uint(1), 10, 10).
		WillReturnRows(rows)

	logs, err := repo.GetOperationLogsByUserID(1, 2, 10)

	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, uint(30), logs[0].ID)
	assert.Equal(t, model.OperationLogActionDelete, logs[0].Action)
	assert.Equal(t, "old.txt", logs[0].TargetName)
	assert.Equal(t, uint(25), logs[1].ID)
	assert.Equal(t, model.OperationLogActionUpload, logs[1].Action)
	assert.Equal(t, "report.pdf", logs[1].TargetName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_GetOperationLogsByUserID_Empty(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `operation_log_models`")).
		WithArgs(uint(1), 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "action", "target_type", "target_id", "target_name", "folder_name", "detail", "created_at"}))

	logs, err := repo.GetOperationLogsByUserID(1, 1, 20)

	assert.NoError(t, err)
	assert.Empty(t, logs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_CountOperationLogsByUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `operation_log_models`")).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	count, err := repo.CountOperationLogsByUserID(1)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_PruneOperationLogs_NotExceedKeep(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	// 日志总量不超过 keep，cutoff 查询查不到记录，不应执行任何 DELETE
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `operation_log_models`")).
		WithArgs(uint(1), 1, 500).
		WillReturnError(gorm.ErrRecordNotFound)

	deleted, err := repo.PruneOperationLogs(1, 500)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_PruneOperationLogs_Prunes(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	// 第 500+1 新的日志 id=60，删除 id < 60 的全部记录
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `operation_log_models`")).
		WithArgs(uint(1), 1, 500).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(60))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `operation_log_models`")).
		WithArgs(uint(1), uint(60)).
		WillReturnResult(sqlmock.NewResult(0, 40))
	mock.ExpectCommit()

	deleted, err := repo.PruneOperationLogs(1, 500)

	assert.NoError(t, err)
	assert.Equal(t, int64(40), deleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_PruneOperationLogs_QueryError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewOperationLogRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `operation_log_models`")).
		WithArgs(uint(1), 1, 500).
		WillReturnError(gorm.ErrInvalidDB)

	deleted, err := repo.PruneOperationLogs(1, 500)

	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}
