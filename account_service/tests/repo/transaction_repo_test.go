package repo_test

import (
	"accountservice/internal/model"
	"accountservice/internal/repo"
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRepoInsertOne(t *testing.T) {
	transactionRepo, err := repo.NewTransactionPostgresRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name                  string
		input                 model.TransactionRequest
		op                    model.Operation
		expectedTransactionId uint
		expectedError         *pgconn.PgError
	}{
		{"First transaction should have id 1", model.TransactionRequest{AccountId: 1, Amount: 100, Currency: "USD"}, model.Invoice, 1, nil},
		{"Second transaction should have id 2", model.TransactionRequest{AccountId: 1, Amount: 100, Currency: "RUB"}, model.Withdraw, 2, nil},
		{"Unexisting accountId should fail", model.TransactionRequest{AccountId: 3, Amount: 100, Currency: "RUB"}, model.Withdraw, 0, &pgconn.PgError{Code: "23503"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gotTransactionId, err := transactionRepo.InsertOne(ctx, tt.input, tt.op)
			if err != nil {
				if tt.expectedError != nil {
					require.ErrorAs(t, err, &tt.expectedError)
				}
			}
			assert.Equal(t, tt.expectedTransactionId, gotTransactionId)
		})
	}
}

func TestTransactionRepoUpdateOne(t *testing.T) {
	transactionRepo, err := repo.NewTransactionPostgresRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name               string
		inputTransactionId uint
		inputStatus        model.Status
	}{
		{"Update status to Error", 1, model.Error},
		{"Update status to Success", 2, model.Success},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			err := transactionRepo.UpdateOne(ctx, tt.inputTransactionId, tt.inputStatus)
			require.NoError(t, err)

			transaction, err := transactionRepo.FindOne(ctx, tt.inputTransactionId)
			require.NoError(t, err)

			assert.Equal(t, tt.inputStatus, transaction.Status)
		})
	}
}
