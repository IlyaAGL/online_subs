// @title Online Subscriptions API
// @version 1.0
// @description REST API for managing user online subscriptions
// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"github.com/agl/online_subs/internal/application/service"
	"github.com/agl/online_subs/internal/infrastructure/repo"
	"github.com/agl/online_subs/internal/presentation/controllers"
	"github.com/agl/online_subs/pkg/bootstrap/connections"
)

func main() {
	db := connections.InitPostgres()
	defer db.Close()

	repo_pg := repo.NewSubsRepo(db)

	service := service.NewSubsService(repo_pg)

	controller := controllers.NewSubsController(service)

	controller.StartServer()
}
