package migrations

import (
	"database/sql"
	"os"

	"github.com/agl/online_subs/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrationsPG(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Log.Error("bootstrap: failed to create pg driver", "error", err)

		return
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		logger.Log.Error("bootstrap: failed to create migrate instance", "error", err)

		return
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Log.Error("bootstrap: migration failed", "error", err)

		return
	}

	logger.Log.Info("Migrations applied successfully")
}
