package main

import (
	"github.com/agl/online_subs/pkg/bootstrap/connections"
	"github.com/agl/online_subs/pkg/bootstrap/migrations"
)

func main() {
	db_pg := connections.InitPostgres()
	defer db_pg.Close()

	migrations.RunMigrationsPG(db_pg)
}
