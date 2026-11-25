package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jshmcz/ChargingHub/internal/app"
	"github.com/jshmcz/ChargingHub/internal/config"
	"github.com/jshmcz/ChargingHub/internal/device"
)

func main() {
	cfg := config.Load()
	db, err := app.OpenDB(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	a := app.New(db, cfg)
	gateway := device.NewGateway(db)
	if err = gateway.Start(":" + cfg.DevicePort); err != nil {
		log.Fatalf("device gateway: %v", err)
	}
	a.SetSender(gateway)

	stop := make(chan struct{})
	go a.RunScheduler(stop)
	go a.RunNotifier(stop)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           a.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("HTTP server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	close(stop)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("shutdown complete")
}
