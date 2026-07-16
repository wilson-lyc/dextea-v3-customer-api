package main

import (
	"log"
	"net/http"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/db"
	"github.com/dextea-v3/dextea-customer/api/internal/demo"
	"github.com/dextea-v3/dextea-customer/api/internal/router"
)

func main() {
	cfg := config.Load()

	// 打开数据库连接（sqlx），供各业务模块的 repository 共用。
	// 若未配置 DATABASE_DSN，db 为 nil，健康接口会显示 db:"disabled"。
	// 接入真实数据库时，需在 main 中匿名导入对应驱动，例如：
	//   import _ "modernc.org/sqlite"
	database, err := db.New(cfg.DatabaseDriver, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("open db error: %v", err)
	}
	if database != nil {
		defer database.Close()
	}

	// 按业务模块组装依赖：repository -> service -> handler。
	repo := demo.NewRepository(database)
	svc := demo.NewService(repo, cfg)
	demoHandler := demo.NewHandler(svc)

	r := router.Setup(cfg, demoHandler)

	addr := ":" + cfg.Port
	log.Printf("%s listening on %s (env=%s)", cfg.ServiceName, addr, cfg.Environment)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
