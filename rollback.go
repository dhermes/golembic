package golembic

import (
	"database/sql"
	"log"
)

func rollbackAndLog(tx *sql.Tx) {
	err := tx.Rollback()
	if err == nil || err == sql.ErrTxDone {
		return
	}

	// TODO: https://github.com/dhermes/golembic/issues/1
	log.Println(err)
}
