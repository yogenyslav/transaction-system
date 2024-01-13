package repo

import (
	"context"
	"fmt"

	"accountservice/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo interface {
	InsertOne(c context.Context, in model.TransactionRequest, op model.Operation) (uint, error)
	UpdateOne(c context.Context, transactionId uint, status model.Status) error
}

type transactionPostgresRepo struct {
	db *pgxpool.Pool
}

func NewTransactionPostgresRepo(db *pgxpool.Pool) (TransactionRepo, error) {
	ctx := context.Background()
	_, err := db.Exec(ctx, fmt.Sprintf(`
		create table if not exists %s(
			id serial primary key,
			fk_account_id int references %s(id),
			amount numeric not null,
			currency text not null,
			operation smallint not null,
			status smallint not null,
			created_at timestamp default current_timestamp
		)
	`, model.TransactionsTable, model.AccountsTable))
	return transactionPostgresRepo{db}, err
}

func (r transactionPostgresRepo) InsertOne(c context.Context, in model.TransactionRequest, op model.Operation) (uint, error) {
	var transactionId uint
	err := r.db.QueryRow(c, fmt.Sprintf(`
		insert into %s(fk_account_id, amount, currency, operation, status)
		values ($1, $2, $3, $4, $5)
		returning id
	`, model.TransactionsTable), in.AccountId, in.Amount, in.Currency, op, model.Created).Scan(&transactionId)
	return transactionId, err
}

func (r transactionPostgresRepo) UpdateOne(c context.Context, transactionId uint, status model.Status) error {
	_, err := r.db.Exec(c, fmt.Sprintf(`
		update %s
		set status=$1
		where id=$2
	`, model.TransactionsTable), status, transactionId)
	return err
}
