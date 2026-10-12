package database

import (
	"context"
	"database/sql"
	"embed"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"io/fs"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, mustSub())
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}

func mustSub() fs.FS {
	f, err := fs.Sub(migrations, "migrations")
	if err != nil {
		panic(err)
	}
	return f
}
