package repo

import (
	"context"
	"errors"
	"time"
	"user-aggregation/internal/models"

	"github.com/google/uuid"
)

type Repo interface {
	Insert(ctx context.Context, u *models.UserInfo) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	UpdateUserInfo(ctx context.Context, userID uuid.UUID, price *int64, end *time.Time) (int64, error)
	List(ctx context.Context) ([]models.UserInfo, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserInfo, error)
	FilterSum(ctx context.Context, userID *uuid.UUID, serviceName *string, start, end *time.Time) (int64, error)
}

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrBadInput   = errors.New("bad input")
	ErrConstraint = errors.New("constraint violation")
)
