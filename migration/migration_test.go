package migration_test

import (
	"testing"
)

func TestMigrate_SkipWithoutDB(t *testing.T) {
	t.Skip("migration runner tested with real Postgres in later steps")
}
