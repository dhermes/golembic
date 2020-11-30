package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/dhermes/golembic/command"
	"github.com/dhermes/golembic/examples"
)

// NOTE: Ensure that
//       * `examples.AllMigrations` satisfies `command.RegisterMigrations`.
var (
	_ command.RegisterMigrations = examples.AllMigrations
)

func main() {
	cmd, err := command.MakeRootCommand(examples.AllMigrations)
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
