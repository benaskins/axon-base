package scan_test

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/benaskins/axon-base/scan"
)

// User is an example domain type used to illustrate scan usage.
type User struct {
	ID   string
	Name string
	Age  int
}

func ExampleRow() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, "postgres://postgres@localhost:5432/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	row := db.QueryRow(ctx, "SELECT id, name, age FROM users WHERE id = $1", "abc")
	user, err := scan.Row(row, func(u *User) []any {
		return []any{&u.ID, &u.Name, &u.Age}
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = user
}

func ExampleRows() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, "postgres://postgres@localhost:5432/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(ctx, "SELECT id, name, age FROM users ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	users, err := scan.Rows(rows, func(u *User) []any {
		return []any{&u.ID, &u.Name, &u.Age}
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = users
}
