package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"accountservice/internal/model"
	"accountservice/internal/repo"
	"accountservice/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type accountController struct {
	transactionClient *service.TransactionClient
	accountRepo       repo.AccountRepo
	transactionRepo   repo.TransactionRepo
}

func NewAccountController(client *service.TransactionClient, ar repo.AccountRepo, tr repo.TransactionRepo) accountController {
	return accountController{
		transactionClient: client,
		accountRepo:       ar,
		transactionRepo:   tr,
	}
}

func (ac accountController) syncBalances(convertedAmount float64, transactionId uint, in model.TransactionRequest, op model.Operation) {
	var (
		ctx           context.Context = context.Background()
		frozenChange  float64         = -1 * convertedAmount
		balanceChange float64         = 0
		status        model.Status
		err           error
	)

	slog.Debug("processing transaction")
	status, err = ac.transactionClient.ProcessTransaction(ctx, transactionId)
	if err != nil || status == model.Error {
		slog.Error("failed to process transaction", slog.Any("error", err))
	} else if err == nil && status == model.Success {
		slog.Debug("processing successfuly completed")
	}

	switch op {
	case model.Invoice:
		if status == model.Success {
			balanceChange += convertedAmount
		}
	case model.Withdraw:
		if status != model.Success {
			balanceChange += convertedAmount
		}
	}

	// TODO: нужно ли дополнительно обрабатывать ошибку при отмене транзакции и сбросе frozen?
	if err := ac.transactionRepo.UpdateOne(ctx, transactionId, status); err != nil {
		slog.Error("failed to update transaction status")
		return
	}

	if err := ac.accountRepo.UpdateOne(ctx, in.AccountId, balanceChange, frozenChange); err != nil {
		slog.Error("failed to update account")
		return
	}
}

func (ac accountController) Invoice(c *fiber.Ctx) error {
	var in model.TransactionRequest
	if err := c.BodyParser(&in); err != nil {
		return model.ErrorResponse{
			Code: http.StatusUnprocessableEntity,
			Msg:  "failed to parse transactionRequest body",
			Err:  err,
		}
	}

	convertedAmount, err := service.Convert(in.Currency, in.Amount)
	if err != nil {
		return model.ErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("%s currency is not supported now", in.Currency),
			Err:  err,
		}
	}

	if err := ac.accountRepo.UpdateOne(c.Context(), in.AccountId, 0, convertedAmount); err != nil {
		return model.ErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  "failed to update account",
			Err:  err,
		}
	}

	transactionId, err := ac.transactionRepo.InsertOne(c.Context(), in, model.Invoice)
	if err != nil {
		return model.ErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  "failed to create invoice transaction",
			Err:  err,
		}
	}

	go ac.syncBalances(convertedAmount, transactionId, in, model.Invoice)

	return c.SendStatus(http.StatusCreated)
}

func (ac accountController) Withdraw(c *fiber.Ctx) error {
	var in model.TransactionRequest
	if err := c.BodyParser(&in); err != nil {
		return model.ErrorResponse{
			Code: http.StatusUnprocessableEntity,
			Msg:  "failed to parse transactionRequest body",
			Err:  err,
		}
	}

	convertedAmount, err := service.Convert(in.Currency, in.Amount)
	if err != nil {
		return model.ErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("%s currency is not supported now", in.Currency),
			Err:  err,
		}
	}

	account, err := ac.accountRepo.FindOne(c.Context(), in.AccountId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrorResponse{
				Code: http.StatusNotFound,
				Msg:  "account record not found",
				Err:  err,
			}
		}
		return model.ErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  "failed to get account",
			Err:  err,
		}
	}

	if account.Balance < convertedAmount {
		return model.ErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  "can't withdraw more than active balance",
		}
	}

	if err := ac.accountRepo.UpdateOne(c.Context(), in.AccountId, -1*convertedAmount, convertedAmount); err != nil {
		return model.ErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  "failed to update account",
			Err:  err,
		}
	}

	transactionId, err := ac.transactionRepo.InsertOne(c.Context(), in, model.Withdraw)
	if err != nil {
		return model.ErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  "failed to create withdraw transaction",
			Err:  err,
		}
	}

	go ac.syncBalances(convertedAmount, transactionId, in, model.Withdraw)

	return c.SendStatus(http.StatusCreated)
}

func (ac accountController) List(c *fiber.Ctx) error {
	accounts, err := ac.accountRepo.FindAll(c.Context())
	if err != nil {
		return model.ErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  "failed to get accounts",
			Err:  err,
		}
	}
	return c.Status(http.StatusOK).JSON(accounts)
}
