package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Exec runs an ExecContext on the provided db/tx.
func Exec(ctx context.Context, db sqlExecer, query string, args ...any) error {
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// ExecRequireRow runs an ExecContext and returns sql.ErrNoRows if it affected 0 rows.
func ExecRequireRow(ctx context.Context, db sqlExecer, query string, args ...any) error {
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ExecOrWrap runs Exec and wraps any error with the provided message.
func ExecOrWrap(ctx context.Context, db sqlExecer, query string, wrapMsg string, args ...any) error {
	if err := Exec(ctx, db, query, args...); err != nil {
		return fmt.Errorf("%s: %w", wrapMsg, err)
	}
	return nil
}

// ExecRequireRowOrWrap runs ExecRequireRow and:
// - returns nil if a row was affected,
// - returns sql.ErrNoRows if 0 rows were affected,
// - wraps any other error with the provided message.
func ExecRequireRowOrWrap(ctx context.Context, db sqlExecer, query string, wrapMsg string, args ...any) error {
	if err := ExecRequireRow(ctx, db, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("%s: %w", wrapMsg, err)
	}
	return nil
}
