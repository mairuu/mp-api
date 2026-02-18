package main

import (
	"github.com/gin-gonic/gin"
	buckethandler "github.com/mairuu/mp-api/internal/features/bucket/handler"
	bucket "github.com/mairuu/mp-api/internal/features/bucket/model"
	bucketservice "github.com/mairuu/mp-api/internal/features/bucket/service"
	mangahandler "github.com/mairuu/mp-api/internal/features/manga/handler"
	manga "github.com/mairuu/mp-api/internal/features/manga/model"
	mangarepo "github.com/mairuu/mp-api/internal/features/manga/repository"
	mangaservice "github.com/mairuu/mp-api/internal/features/manga/service"
	userhandler "github.com/mairuu/mp-api/internal/features/user/handler"
	user "github.com/mairuu/mp-api/internal/features/user/model"
	userrepo "github.com/mairuu/mp-api/internal/features/user/repository"
	userservice "github.com/mairuu/mp-api/internal/features/user/service"
	"github.com/mairuu/mp-api/internal/platform/authentication"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/config"
	"github.com/mairuu/mp-api/internal/platform/database"
	"github.com/mairuu/mp-api/internal/platform/logging"
	"github.com/mairuu/mp-api/internal/platform/storage"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
	"github.com/mairuu/mp-api/internal/platform/transport/http/handler"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	log := logging.New(cfg.App.LogLevel)

	db, err := database.NewClient(&cfg.DB, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}

	b, err := storage.NewBucket(&cfg.Storage)
	if err != nil {
		log.Error("failed to create bucket", "error", err)
		panic(err)
	}

	enforcer, err := authorization.NewEnforcer()
	if err != nil {
		log.Error("failed to initialize authorization enforcer", "error", err)
		panic(err)
	}

	err = enforcer.AddPolicies(
		bucket.AllPolicies(),
		user.AllPolicies(),
		manga.AllPolicies(),
	)
	if err != nil {
		log.Error("failed to add policies to enforcer", "error", err)
		panic(err)
	}

	tokenService := authentication.NewTokenService(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)
	userRepo := userrepo.NewGormRepository(db)
	mangaRepo := mangarepo.NewGormRepository(db)

	bucketService := bucketservice.NewService(enforcer, b)
	userService := userservice.NewService(userRepo, tokenService, enforcer)
	mangaService := mangaservice.NewService(log, mangaRepo, enforcer, b)

	r := gin.New()
	r.SetTrustedProxies(nil)

	r.Use(gin.Recovery())
	r.Use(middleware.CORS("*")) // todo: make configurable
	r.Use(middleware.TraceID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Auth(tokenService))

	router := httptransport.NewRouter(r, []httptransport.Handler{
		handler.NewHealthHandler(log),
		buckethandler.NewBucketHandler(log, bucketService),
		userhandler.NewUserHandler(log, userService),
		mangahandler.NewHandler(log, mangaService),
	})
	router.RegisterRoutes()

	// todo: graceful shutdown
	log.Info("starting server", "addr", cfg.HTTP.Addr)
	if err := r.Run(cfg.HTTP.Addr); err != nil {
		log.Error("failed to run server", "error", err)
	}
}
