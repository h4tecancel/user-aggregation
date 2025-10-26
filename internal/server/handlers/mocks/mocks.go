// internal/server/handlers/mocks/mocks.go
package mocks

import (
	"context"
	"time"
	"user-aggregation/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type RepoMock struct {
	mock.Mock
}

func (m *RepoMock) Insert(ctx context.Context, u *models.UserInfo) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *RepoMock) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RepoMock) UpdateUserInfo(ctx context.Context, userID uuid.UUID, price *int64, end *time.Time) (int64, error) {
	args := m.Called(ctx, userID, price, end)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RepoMock) List(ctx context.Context) ([]models.UserInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.UserInfo), args.Error(1)
}

func (m *RepoMock) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.UserInfo), args.Error(1)
}

func (m *RepoMock) FilterSum(ctx context.Context, userID *uuid.UUID, serviceName *string, start, end *time.Time) (int64, error) {
	args := m.Called(ctx, userID, serviceName, start, end)
	return args.Get(0).(int64), args.Error(1)
}
