package main

import (
	"log"
	"net/http"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/router"
)

func main() {
	cfg := config.Load()

	r := router.Setup(cfg)

	addr := ":" + cfg.Port
	log.Printf("dextea-customer-api listening on %s (env=%s)", addr, cfg.Environment)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
