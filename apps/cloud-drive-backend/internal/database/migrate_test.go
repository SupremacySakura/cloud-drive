package database

import (
	"errors"
	"fmt"
	"testing"

	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/utils"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUploadIdentityBackfills_PreservesOneCanonicalTaskAndSeparatesDuplicates(t *testing.T) {
	tasks := []model.UploadTask{
		{
			ID: 1, UserID: 7, FolderID: 9, FileName: " report.pdf ", FileHash: "ABCDEF",
			FileSize: 1024, ChunkSize: 512, TotalChunks: 2, FileType: "application/pdf",
		},
		{
			ID: 2, UserID: 7, FolderID: 9, FileName: " report.pdf ", FileHash: "ABCDEF",
			FileSize: 1024, ChunkSize: 512, TotalChunks: 2, FileType: "application/pdf",
		},
	}

	backfills, err := buildUploadIdentityBackfills(tasks)

	require.NoError(t, err)
	require.Len(t, backfills, 2)
	assert.NotEmpty(t, backfills[0].IdempotencyKey)
	assert.NotEmpty(t, backfills[0].RequestHash)
	assert.NotEqual(t, backfills[0].IdempotencyKey, backfills[1].IdempotencyKey)
	assert.Equal(t, backfills[0].RequestHash, backfills[1].RequestHash)
}

func TestBuildUploadIdentityBackfills_DoesNotReplaceAnExistingKey(t *testing.T) {
	tasks := []model.UploadTask{
		{
			ID: 1, UserID: 7, FileName: "report.pdf", FileHash: "abcdef",
			FileSize: 1024, ChunkSize: 512, TotalChunks: 2, FileType: "application/pdf",
			IdempotencyKey: "already-assigned",
		},
	}

	backfills, err := buildUploadIdentityBackfills(tasks)

	require.NoError(t, err)
	require.Len(t, backfills, 1)
	assert.Equal(t, "already-assigned", backfills[0].IdempotencyKey)
	assert.NotEmpty(t, backfills[0].RequestHash)
}

func TestBuildUploadIdentityBackfills_SkipsTasksWhoseIdentityIsAlreadyBackfilled(t *testing.T) {
	task := model.UploadTask{
		ID: 1, UserID: 7, FileName: "report.pdf", FileHash: "abcdef",
		FileSize: 5, ChunkSize: 5, TotalChunks: 1, FileType: "text/plain",
	}
	key, requestHash, _, err := utils.BuildUploadIdentity(&task)
	require.NoError(t, err)
	task.IdempotencyKey = key
	task.RequestHash = requestHash

	backfills, err := buildUploadIdentityBackfills([]model.UploadTask{task})

	require.NoError(t, err)
	assert.Empty(t, backfills)
}

func TestNormalizeUploadIdentityIndexError_IgnoresDuplicateIndex(t *testing.T) {
	err := fmt.Errorf("create index: %w", &mysqlDriver.MySQLError{
		Number:  1061,
		Message: "Duplicate key name 'uniq_upload_tasks_idempotency_key'",
	})

	assert.NoError(t, normalizeUploadIdentityIndexError(err))
}

func TestNormalizeUploadIdentityIndexError_PreservesOtherErrors(t *testing.T) {
	want := errors.New("connection lost")

	assert.ErrorIs(t, normalizeUploadIdentityIndexError(want), want)
}
