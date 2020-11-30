package main

import (
	"context"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/examples"
	"github.com/dhermes/golembic/postgres"
)

func mustEnvVar(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Required environment variable %q is not set.", name)
	}
	return value
}

func deferredClose(manager *golembic.Manager) {
	err := manager.CloseConnectionPool()
	if err != nil {
		log.Fatal(err)
	}

	return
}

func run() error {
	sqlDirectory := mustEnvVar("GOLEMBIC_SQL_DIR")
	migrations, err := examples.AllMigrations(sqlDirectory)
	if err != nil {
		return err
	}

	provider, err := postgres.New(
		postgres.OptHost(mustEnvVar("DB_HOST")),
		postgres.OptPort(mustEnvVar("DB_PORT")),
		postgres.OptDatabase(mustEnvVar("DB_NAME")),
		postgres.OptUsername(mustEnvVar("DB_USER")),
		postgres.OptPassword(mustEnvVar("PGPASSWORD")),
		postgres.OptSSLMode(mustEnvVar("DB_SSLMODE")),
	)
	if err != nil {
		return err
	}

	manager, err := golembic.NewManager(
		golembic.OptDevelopmentMode(true),
		golembic.OptManagerProvider(provider),
		golembic.OptManagerSequence(migrations),
	)
	if err != nil {
		return err
	}
	defer deferredClose(manager)

	ctx := context.Background()
	return manager.Up(ctx)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
