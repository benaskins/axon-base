package errors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	axonerrors "github.com/benaskins/axon-base/errors"
)

func TestWrapError_nilPassthrough(t *testing.T) {
	if axonerrors.WrapError("op", nil) != nil {
		t.Fatal("expected nil for nil error")
	}
}

func TestWrapError_addsContext(t *testing.T) {
	base := errors.New("base error")
	wrapped := axonerrors.WrapError("myop", base)
	if wrapped == nil {
		t.Fatal("expected non-nil error")
	}
	want := "myop: base error"
	if wrapped.Error() != want {
		t.Fatalf("got %q, want %q", wrapped.Error(), want)
	}
	if !errors.Is(wrapped, base) {
		t.Fatal("wrapped error should unwrap to base")
	}
}

func TestIsNotFoundError(t *testing.T) {
	if axonerrors.IsNotFoundError(nil) {
		t.Fatal("nil should not be not-found")
	}
	if !axonerrors.IsNotFoundError(axonerrors.ErrNotFound) {
		t.Fatal("ErrNotFound should be detected")
	}
	wrapped := fmt.Errorf("get: %w", axonerrors.ErrNotFound)
	if !axonerrors.IsNotFoundError(wrapped) {
		t.Fatal("wrapped ErrNotFound should be detected")
	}
	if axonerrors.IsNotFoundError(errors.New("other")) {
		t.Fatal("random error should not be not-found")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if axonerrors.IsUniqueViolation(nil) {
		t.Fatal("nil should not be unique violation")
	}
	pgErr := &pgconn.PgError{Code: "23505"}
	if !axonerrors.IsUniqueViolation(pgErr) {
		t.Fatal("pg unique violation should be detected")
	}
	wrapped := fmt.Errorf("create: %w", pgErr)
	if !axonerrors.IsUniqueViolation(wrapped) {
		t.Fatal("wrapped pg unique violation should be detected")
	}
	pgOther := &pgconn.PgError{Code: "23503"}
	if axonerrors.IsUniqueViolation(pgOther) {
		t.Fatal("other pg error should not be unique violation")
	}
}
