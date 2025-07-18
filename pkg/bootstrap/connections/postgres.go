package connections

import (
	"database/sql"
	"os"
	"time"

	"github.com/agl/online_subs/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitPostgres() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log.Error("bootstrap: failed to connect to DB", "error", err)

		return nil
	}

	for i := range 10 {
		err = db.Ping()
		if err == nil {
			break
		}

		logger.Log.Info("bootstrap: retrying DB connection...", "attempt", i+1)

		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Log.Error("bootstrap: could not ping DB", "error", err)

		return nil
	}

	return db
}
