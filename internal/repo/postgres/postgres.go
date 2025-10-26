package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"user-aggregation/internal/models"
	"user-aggregation/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string, maxConns int32) (*Repo, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("repo: parse db url: %w", err)
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("repo: connect db: %w", err)
	}
	return &Repo{pool: pool}, nil
}

func NewFromPool(pool *pgxpool.Pool) *Repo {
	if pool == nil {
		panic("repo: nil pgxpool")
	}
	return &Repo{pool: pool}
}

func (p *Repo) Ping(ctx context.Context) error {
	if p == nil || p.pool == nil {
		return errors.New("repo: nil pool")
	}
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("repo: ping: %w", err)
	}
	return nil
}

func (p *Repo) Close() {
	if p != nil && p.pool != nil {
		p.pool.Close()
	}
}

func (p *Repo) Insert(ctx context.Context, u *models.UserInfo) error {
	if u == nil {
		return errors.Join(repo.ErrBadInput, errors.New("nil user info"))
	}
	const q = `
			INSERT INTO user_info (service_name, price, user_id, start_date, end_date)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id, service_name, start_date) DO UPDATE
			SET price = EXCLUDED.price,
    		end_date = EXCLUDED.end_date`
	ct, err := p.pool.Exec(ctx, q, u.ServiceName, u.Price, u.UserID, u.StartDate, u.EndDate)
	if err != nil {
		return fmt.Errorf("repo: insert user_info: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return errors.Join(repo.ErrConflict, errors.New("no rows affected"))
	}
	return nil
}

func (p *Repo) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	const q = `DELETE FROM user_info WHERE user_id = $1`
	ct, err := p.pool.Exec(ctx, q, userID)
	if err != nil {
		return 0, fmt.Errorf("repo: delete by user_id: %w", err)
	}
	n := ct.RowsAffected()
	if n == 0 {
		return 0, repo.ErrNotFound
	}
	return n, nil
}

func (p *Repo) UpdateUserInfo(ctx context.Context, userID uuid.UUID, price *int64, end *time.Time) (int64, error) {
	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)

	if price != nil {
		args = append(args, *price)
		sets = append(sets, fmt.Sprintf("price = $%d", len(args)))
	}
	if end != nil {
		args = append(args, *end)
		sets = append(sets, fmt.Sprintf("end_date = $%d", len(args)))
	}

	if len(sets) == 0 {
		return 0, fmt.Errorf("repo: patch user_info: no fields to update")
	}

	args = append(args, userID)
	q := fmt.Sprintf(`
		UPDATE user_info
		SET %s
		WHERE user_id = $%d
	`, strings.Join(sets, ", "), len(args))

	ct, err := p.pool.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("repo: patch user_info: %w", err)
	}

	n := ct.RowsAffected()
	if n == 0 {
		return 0, repo.ErrNotFound
	}
	return n, nil
}

func (p *Repo) List(ctx context.Context) ([]models.UserInfo, error) {
	const q = `
			SELECT service_name, price, user_id, start_date, end_date
			FROM user_info
			ORDER BY user_id, service_name, start_date`
	rows, err := p.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repo: list user_info: %w", err)
	}
	defer rows.Close()

	var out []models.UserInfo
	for rows.Next() {
		u, err := scanUserInfo(rows)
		if err != nil {
			return nil, fmt.Errorf("repo: scan user_info: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: iterate user_info: %w", err)
	}
	return out, nil
}

func (p *Repo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserInfo, error) {
	const q = `
			SELECT service_name, price, user_id, start_date, end_date
			FROM user_info
			WHERE user_id = $1
			ORDER BY service_name, start_date`
	rows, err := p.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repo: select by user_id: %w", err)
	}
	defer rows.Close()

	var out []models.UserInfo
	for rows.Next() {
		u, err := scanUserInfo(rows)
		if err != nil {
			return nil, fmt.Errorf("repo: scan by user_id: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: iterate by user_id: %w", err)
	}
	return out, nil
}

func (p *Repo) FilterSum(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	start, end *time.Time,
) (int64, error) {
	conds := make([]string, 0, 5)
	args := make([]any, 0, 5)

	conds = append(conds, "1=1")

	if start != nil && !start.IsZero() {
		args = append(args, *start)
		conds = append(conds, fmt.Sprintf("COALESCE(end_date, 'infinity') >= $%d", len(args)))
	}
	if end != nil && !end.IsZero() {
		args = append(args, *end)
		conds = append(conds, fmt.Sprintf("start_date <= $%d", len(args)))
	}
	if userID != nil && *userID != uuid.Nil {
		args = append(args, *userID)
		conds = append(conds, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if serviceName != nil && *serviceName != "" {
		args = append(args, *serviceName)
		conds = append(conds, fmt.Sprintf("service_name = $%d", len(args)))
	}

	q := "SELECT COALESCE(SUM(price), 0) FROM user_info WHERE " + strings.Join(conds, " AND ")

	var total int64
	if err := p.pool.QueryRow(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("repo: filter sum: %w", err)
	}
	return total, nil
}

func (p *Repo) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("repo: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() 

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repo: commit tx: %w", err)
	}
	return nil
}

func scanUserInfo(r pgx.Rows) (models.UserInfo, error) {
	var u models.UserInfo
	if err := r.Scan(&u.ServiceName, &u.Price, &u.UserID, &u.StartDate, &u.EndDate); err != nil {
		return models.UserInfo{}, err
	}
	return u, nil
}
