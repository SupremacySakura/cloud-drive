package service

import (
	"errors"
	"testing"

	"cloud-drive-backend/internal/model"
	"cloud-drive-backend/internal/utils"

	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct {
	user *model.UserModel
	err  error
}

func (m *mockUserRepo) GetUserByName(username string) (*model.UserModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil || m.user.Username != username {
		return nil, errors.New("用户不存在")
	}
	return m.user, nil
}

func (m *mockUserRepo) Create(user *model.UserModel) error {
	return nil
}

func (m *mockUserRepo) GetUserByID(userID uint) (*model.UserModel, error) {
	return nil, nil
}

func TestValidateUser_UserNotFound(t *testing.T) {
	mockRepo := &mockUserRepo{err: errors.New("用户不存在")}
	svc := &authService{UserRepo: mockRepo}

	user, err := svc.ValidateUser("nonexistent", "password")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "用户不存在")
}

func TestValidateUser_Success(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("correctpassword")
	mockRepo := &mockUserRepo{
		user: &model.UserModel{ID: 1, Username: "testuser", PasswordHash: hashedPassword},
	}
	svc := &authService{UserRepo: mockRepo}

	user, err := svc.ValidateUser("testuser", "correctpassword")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
}

func TestValidateUser_WrongPassword(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("correctpassword")
	mockRepo := &mockUserRepo{
		user: &model.UserModel{ID: 1, Username: "testuser", PasswordHash: hashedPassword},
	}
	svc := &authService{UserRepo: mockRepo}

	user, err := svc.ValidateUser("testuser", "wrongpassword")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "密码错误")
}
