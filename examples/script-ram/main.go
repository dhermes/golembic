package main

import (
	"context"
	"log"
	"os"

	_ "github.com/proullon/ramsql/driver"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/examples"
	"github.com/dhermes/golembic/ramsql"
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

func main() {
	sqlDirectory := mustEnvVar("GOLEMBIC_SQL_DIR")
	migrations, err := examples.AllMigrations(sqlDirectory)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := ramsql.New(
		ramsql.OptDataSourceName("ramsql-script"),
	)
	if err != nil {
		log.Fatal(err)
	}

	manager, err := golembic.NewManager(
		golembic.OptManagerProvider(provider),
		golembic.OptManagerSequence(migrations),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer deferredClose(manager)

	ctx := context.Background()
	err = manager.Up(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
