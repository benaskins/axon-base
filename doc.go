// Package base provides PostgreSQL foundation primitives: connection pooling,
// repository interfaces, goose migrations, and row scanning helpers.
//
// Subpackages: repository, pool, migrate.
//
// Class: primitive
// UseWhen: Any service needing relational data access. Pool for connections, migration.Run for embedded SQL, scan.Row/scan.Rows with RowMapper for type-safe CRUD. Domain stores use explicit SQL, not an ORM.
package base
