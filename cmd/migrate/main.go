package main

import (
	"github.com/mairuu/mp-api/internal/persistence/models"
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
		&models.UserDB{},
		&models.RefreshTokenDB{},
		&models.MangaDB{},
		&models.CoverArtDB{},
		&models.ChapterDB{},
		&models.ChapterPageDB{},
		&models.LibraryMangaDB{},
		&models.HistoryDB{},
	}
	if err := db.AutoMigrate(allModels...); err != nil {
		log.Error("failed to migrate database", "error", err)
		panic(err)
	}

	log.Info("database migration completed successfully")
}
