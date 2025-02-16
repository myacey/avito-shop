package main

import (
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/myacey/avito-shop/internal/backconfig"
	"github.com/myacey/avito-shop/internal/controller"
	"github.com/myacey/avito-shop/internal/hasher"
	"github.com/myacey/avito-shop/internal/jwttoken"
	"github.com/myacey/avito-shop/internal/repository/postgresrepo"
	"github.com/myacey/avito-shop/internal/repository/redisrepo"
	"github.com/myacey/avito-shop/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	gin.SetMode(gin.ReleaseMode)
	debug.SetGCPercent(400)

	cfg, err := backconfig.LoadConfig()
	if err != nil {
		panic(err)
	}

	psqlQueries, dbConn, err := postgresrepo.ConfiurePostgres(cfg)
	if err != nil {
		panic(err)
	}
	defer dbConn.Close()
	dbConn.SetMaxOpenConns(570)
	dbConn.SetMaxIdleConns(570)

	usrRepo := postgresrepo.NewPostgresUserRepo(psqlQueries)
	inventoryRepo := postgresrepo.NewPostgresInventoryRepo(psqlQueries)
	trxRepo := postgresrepo.NewPostgresTransferRepo(psqlQueries)
	storeRepo := postgresrepo.NewPostgresStoreRepo(psqlQueries)

	jwtSecretKey := "lovushka_jokera"
	if cfg.JWTSecretKey != "" {
		jwtSecretKey = cfg.JWTSecretKey
	}
	tokenMaker := jwttoken.CreateTokenMaker([]byte(jwtSecretKey), 24*time.Hour)
	redisConn, err := redisrepo.ConfigureRedisClient(&cfg)
	if err != nil {
		panic(err)
	}
	sessionRepo := redisrepo.NewRedisSessionRepo(redisConn)

	srv := service.NewService(dbConn, usrRepo, trxRepo, inventoryRepo, storeRepo, sessionRepo, tokenMaker, &hasher.BcryptHasher{})

	handler := controller.NewController(srv)

	r := gin.New()
	pprof.Register(r)
	r.POST("/api/auth", handler.Authorize)

	r.Use(handler.AuthMiddleware())
	r.GET("/api/info", handler.GetFullUserInfo)
	r.POST("/api/sendCoin", handler.SendCoins)
	r.GET("/api/buy/:item", handler.BuyItem)

	log.Printf("start listening on port :%s", cfg.ServerPort)
	if err = r.Run(":" + cfg.ServerPort); err != nil {
		panic(err)
	}
}
