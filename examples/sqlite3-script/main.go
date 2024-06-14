package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/examples"
	"github.com/dhermes/golembic/sqlite3"
)

func requireEnvVar(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		err := fmt.Errorf("Required environment variable %q is not set", name)
		return "", err
	}
	return value, nil
}

func deferredClose(manager *golembic.Manager, err error) error {
	if manager == nil {
		return err
	}

	closeErr := manager.CloseConnectionPool()
	if err == nil {
		return closeErr
	}
	if closeErr == nil {
		return err
	}

	return fmt.Errorf("%w; %v", err, closeErr)
}

type config struct {
	GolembicSQLDir   string
	GolembicSQLiteDB string
}

func (c *config) Resolve() (err error) {
	c.GolembicSQLDir, err = requireEnvVar("GOLEMBIC_SQL_DIR")
	if err != nil {
		return
	}

	c.GolembicSQLiteDB, err = requireEnvVar("GOLEMBIC_SQLITE_DB")
	if err != nil {
		return
	}

	return
}

func run() (err error) {
	var manager *golembic.Manager
	defer func() {
		err = deferredClose(manager, err)
	}()

	c := config{}
	err = c.Resolve()
	if err != nil {
		return
	}

	migrations, err := examples.AllMigrations(c.GolembicSQLDir, "sqlite3")
	if err != nil {
		return
	}

	sqliteDBFull, err := filepath.Abs(c.GolembicSQLiteDB)
	if err != nil {
		return
	}

	dsn := fmt.Sprintf("file:%s?cache=shared", sqliteDBFull)
	provider, err := sqlite3.New(
		sqlite3.OptDataSourceName(dsn),
	)
	if err != nil {
		return
	}

	manager, err = golembic.NewManager(
		golembic.OptDevelopmentMode(true),
		golembic.OptManagerProvider(provider),
		golembic.OptManagerSequence(migrations),
	)
	if err != nil {
		return
	}

	ctx := context.Background()
	err = manager.Up(ctx)
	return
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
