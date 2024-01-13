package main

import (
	"accountservice/internal/api/router"
	"accountservice/internal/config"
	"accountservice/internal/database"
	"accountservice/internal/logging"
	"accountservice/internal/service"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustNewConfig(".env")
	logger := logging.MustNewLogger("main")
	slog.SetDefault(logger)

	transactionClient := service.MustNewTransactionClient(cfg).WithQueue("", false, false).WithConsumer("")
	defer transactionClient.Close()

	db := database.MustNewPostgres(cfg, 3)
	defer db.Close()

	app := router.MustNewApp(transactionClient, db)
	defer app.Shutdown()
	go func() {
		slog.Info("started listening", slog.Int("port", cfg.Server.Port))
		if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
			slog.Error("failed to start server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	_ = <-ch
	slog.Info("shutting down the app")
}
