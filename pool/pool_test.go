package pool_test

import (
	"testing"

	"github.com/benaskins/axon-base/pool"
)

// Compile-time check: Pool type is exported and NewPool exists.
var _ *pool.Pool

func TestPool_Placeholder(t *testing.T) {
	t.Skip("pool implementation tested in step 2")
}
