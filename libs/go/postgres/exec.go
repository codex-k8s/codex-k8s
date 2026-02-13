package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type execer interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

// Exec runs Exec on the provided db/tx.
func Exec(ctx context.Context, db execer, query string, args ...any) error {
	_, err := db.Exec(ctx, query, args...)
	return err
}

// ExecRequireRow runs Exec and returns pgx.ErrNoRows if it affected 0 rows.
func ExecRequireRow(ctx context.Context, db execer, query string, args ...any) error {
	res, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	n := res.RowsAffected()
	if n == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ExecOrWrap runs Exec and wraps any error with the provided message.
func ExecOrWrap(ctx context.Context, db execer, query string, wrapMsg string, args ...any) error {
	if err := Exec(ctx, db, query, args...); err != nil {
		return fmt.Errorf("%s: %w", wrapMsg, err)
	}
	return nil
}

// ExecRequireRowOrWrap runs ExecRequireRow and:
// - returns nil if a row was affected,
// - returns pgx.ErrNoRows if 0 rows were affected,
// - wraps any other error with the provided message.
func ExecRequireRowOrWrap(ctx context.Context, db execer, query string, wrapMsg string, args ...any) error {
	if err := ExecRequireRow(ctx, db, query, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("%s: %w", wrapMsg, err)
	}
	return nil
}
