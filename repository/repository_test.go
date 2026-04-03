package repository_test

import (
	"testing"

	"github.com/benaskins/axon-base/repository"
)

// Compile-time check: Repository interface is defined.
var _ repository.Repository[string]

func TestRepositoryInterface(t *testing.T) {
	t.Skip("concrete implementation tested in later steps")
}
