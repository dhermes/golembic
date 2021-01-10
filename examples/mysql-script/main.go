package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/examples"
	"github.com/dhermes/golembic/mysql"
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
	GolembicSQLDir string
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
}

func (c *config) Resolve() (err error) {
	c.GolembicSQLDir, err = requireEnvVar("GOLEMBIC_SQL_DIR")
	if err != nil {
		return
	}

	c.DBHost, err = requireEnvVar("DB_HOST")
	if err != nil {
		return
	}

	c.DBPort, err = requireEnvVar("DB_PORT")
	if err != nil {
		return
	}

	c.DBName, err = requireEnvVar("DB_NAME")
	if err != nil {
		return
	}

	c.DBUser, err = requireEnvVar("DB_USER")
	if err != nil {
		return
	}

	c.DBPassword, err = requireEnvVar("DB_PASSWORD")
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

	migrations, err := examples.AllMigrations(c.GolembicSQLDir, "mysql")
	if err != nil {
		return
	}

	port, err := strconv.Atoi(c.DBPort)
	if err != nil {
		return
	}

	provider, err := mysql.New(
		mysql.OptNet("tcp"),
		mysql.OptHostPort(c.DBHost, port),
		mysql.OptDBName(c.DBName),
		mysql.OptUser(c.DBUser),
		mysql.OptPassword(c.DBPassword),
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
