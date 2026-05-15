package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)


const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

var ErrRecordNotFound = pgx.ErrNoRows

var ErrUniqueViolation = &pgconn.PgError{
	Code: UniqueViolation,
}

// ErrorCode返回错误的SQLState代码，如果错误不是pgconn.PgError类型，则返回空字符串。
func ErrorCode(err error) string {
	var pgErr *pgconn.PgError 
	if errors.As(err, &pgErr) { //如果错误是pgconn.PgError类型，则返回错误码
		return pgErr.Code
	}
	return ""
}