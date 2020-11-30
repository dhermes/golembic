package main

import (
	"context"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/examples"
	"github.com/dhermes/golembic/mysql"
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
	opt := examples.OptAddUsersEmailFile("0005_add_users_email_index_lock_none.sql")
	migrations, err := examples.AllMigrationsWithOptions(sqlDirectory, opt)
	if err != nil {
		return err
	}

	portStr := mustEnvVar("DB_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	provider, err := mysql.New(
		mysql.OptNet("tcp"),
		mysql.OptHostPort(mustEnvVar("DB_HOST"), port),
		mysql.OptDBName(mustEnvVar("DB_NAME")),
		mysql.OptUser(mustEnvVar("DB_USER")),
		mysql.OptPassword(mustEnvVar("DB_PASSWORD")),
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
