package router

import (
	"errors"
	"log/slog"
	"net/http"

	"accountservice/internal/api/controller"
	"accountservice/internal/errs"
	"accountservice/internal/model"
	"accountservice/internal/repo"
	"accountservice/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MustNewApp(transactionClient *service.TransactionClient, db *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Transaction System",
		ErrorHandler: func(c *fiber.Ctx, e error) error {
			err := &model.ErrorResponse{}
			if !errors.As(e, err) {
				c.Set("Content-Type", "text/plain")
				return c.Status(http.StatusInternalServerError).SendString(e.Error())
			}
			slog.Error(err.Msg, slog.Any("error", err.Err))
			return c.Status(http.StatusInternalServerError).JSON(err)
		},
	})

	SetupMiddlewares(app)
	if err := SetupRoutes(app, transactionClient, db); err != nil {
		panic(err)
	}

	return app
}

func SetupMiddlewares(app *fiber.App) {
	app.Use(logger.New())
	app.Use(recover.New())
}

func SetupRoutes(app *fiber.App, transactionClient *service.TransactionClient, db *pgxpool.Pool) error {
	api := app.Group("/api")

	err := make(map[string]error, 2)
	accountRepo, e := repo.NewAccountPostgresRepo(db)
	err["account"] = e
	transactionRepo, e := repo.NewTransactionPostgresRepo(db)
	err["transaction"] = e

	success := true
	for repoName, e := range err {
		if e != nil {
			slog.Error(e.Error(), slog.String("repoName", repoName))
			success = false
		}
	}
	if !success {
		return errs.ErrRepoCreate
	}

	accountController := controller.NewAccountController(transactionClient, accountRepo, transactionRepo)
	accounts := api.Group("/accounts")
	accounts.Post("/invoice", accountController.Invoice)
	accounts.Post("/withdraw", accountController.Withdraw)
	accounts.Get("/list", accountController.List)

	return nil
}
