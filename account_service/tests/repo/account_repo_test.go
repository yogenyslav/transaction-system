package repo_test

import (
	"accountservice/internal/config"
	"accountservice/internal/database"
	"accountservice/internal/model"
	"accountservice/internal/repo"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cfg *config.Config
	db  *pgxpool.Pool
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	cfg = config.MustNewConfig("../../.env").WithDbHost("localhost")
	db = database.MustNewPostgres(cfg, 3)
	defer func() {
		_, _ = db.Exec(context.Background(), `
			drop table if exists transactions; 
			drop table if exists accounts;
		`)
		db.Close()
	}()

	return m.Run(), nil
}

func TestAccountRepoInsertOne(t *testing.T) {
	accountRepo, err := repo.NewAccountPostgresRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name              string
		expectedAccountId uint
	}{
		{"First account should have id 1", 1},
		{"Second account should have id 2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gotAccountId, err := accountRepo.InsertOne(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAccountId, gotAccountId)
		})
	}
}

func TestAccountRepoFindOne(t *testing.T) {
	accountRepo, err := repo.NewAccountPostgresRepo(db)
	require.NoError(t, err)

	var tests = []struct {
		name            string
		inputAccountId  uint
		expectedAccount model.Account
		expectedError   error
	}{
		{"Trying to find unexisting id should return ErrNoRows", 9999, model.Account{}, pgx.ErrNoRows},
		{"Selecting with id 1 should return account with id 1", 1, model.Account{Id: 1}, nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			gotAccount, err := accountRepo.FindOne(ctx, tt.inputAccountId)
			require.ErrorIs(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedAccount.Id, gotAccount.Id)
		})
	}
}

func TestAccountRepoFindAll(t *testing.T) {
	accountRepo, err := repo.NewAccountPostgresRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name             string
		expectedAccounts []model.Account
	}{
		{"Selecting all accounts should return two accounts", []model.Account{
			{Id: 1, Balance: 0, Frozen: 0},
			{Id: 2, Balance: 0, Frozen: 0},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gotAccounts, err := accountRepo.FindAll(ctx)
			require.NoError(t, err)
			assert.Equal(t, len(tt.expectedAccounts), len(gotAccounts))
		})
	}
}

func TestAccountRepoUpdateOne(t *testing.T) {
	accountRepo, err := repo.NewAccountPostgresRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name           string
		inputAccountId uint
		inputBalance   float64
		inputFrozen    float64
	}{
		{"Balance diff should be +100, frozen diff should be +100", 1, 100, 100},
		{"Balance diff should be 0, frozen diff should be +100", 1, 0, 100},
		{"Balance diff should be 0, frozen diff should be -100", 1, 0, -100},
		{"Balance diff should be -100, frozen diff should be 0", 1, -100, 0},
		{"Balance diff should be +100, frozen diff should be 0", 1, +100, 0},
		{"Balance diff should be -100, frozen diff should be -100", 1, -100, -100},
		{"Balance diff should be 0, frozen diff should be 0", 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			oldAccount, err := accountRepo.FindOne(ctx, tt.inputAccountId)
			require.NoError(t, err)

			err = accountRepo.UpdateOne(ctx, tt.inputAccountId, tt.inputBalance, tt.inputFrozen)
			require.NoError(t, err)

			gotAccount, err := accountRepo.FindOne(ctx, tt.inputAccountId)
			require.NoError(t, err)

			assert.Equal(t, oldAccount.Balance+tt.inputBalance, gotAccount.Balance)
			assert.Equal(t, oldAccount.Frozen+tt.inputFrozen, gotAccount.Frozen)
		})
	}
}
