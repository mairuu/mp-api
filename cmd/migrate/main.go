package main

import (
	mangarepo "github.com/mairuu/mp-api/internal/features/manga/repository"
	userrepo "github.com/mairuu/mp-api/internal/features/user/repository"
	"github.com/mairuu/mp-api/internal/platform/config"
	"github.com/mairuu/mp-api/internal/platform/database"
	"github.com/mairuu/mp-api/internal/platform/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logging.New(cfg.App.LogLevel)

	db, err := database.NewClient(&cfg.DB, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}

	allModels := []any{
		&userrepo.UserDB{},
		&mangarepo.MangaDB{},
		&mangarepo.CoverArtDB{},
		&mangarepo.ChapterDB{},
		&mangarepo.PageDB{},
	}
	if err := db.AutoMigrate(allModels...); err != nil {
		log.Error("failed to migrate database", "error", err)
		panic(err)
	}

	log.Info("database migration completed successfully")
}
