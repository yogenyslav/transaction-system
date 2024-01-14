package repo

import (
	"context"
	"fmt"

	"accountservice/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepo interface {
	InsertOne(c context.Context) (uint, error)
	FindOne(c context.Context, accountId uint) (model.Account, error)
	FindAll(c context.Context) ([]model.Account, error)
	UpdateOne(c context.Context, accountId uint, balanceChange, frozenChange float64) error
}

type accountPostgresRepo struct {
	db *pgxpool.Pool
}

func NewAccountPostgresRepo(db *pgxpool.Pool) (AccountRepo, error) {
	ctx := context.Background()
	_, err := db.Exec(ctx, fmt.Sprintf(`
		create table if not exists %s(
			id serial primary key,
			balance numeric not null,
			frozen numeric not null,
			created_at timestamp default current_timestamp,
			updated_at timestamp default current_timestamp
		)
	`, model.AccountsTable))
	return accountPostgresRepo{db}, err
}

func (r accountPostgresRepo) InsertOne(c context.Context) (uint, error) {
	var accountId uint
	err := r.db.QueryRow(c, fmt.Sprintf(`
		insert into %s(balance, frozen)
		values (0, 0)
		returning id
	`, model.AccountsTable)).Scan(&accountId)
	return accountId, err
}

func (r accountPostgresRepo) FindOne(c context.Context, accountId uint) (model.Account, error) {
	var a model.Account
	err := r.db.QueryRow(c, fmt.Sprintf(`
		select * from %s
		where id = $1
	`, model.AccountsTable), accountId).Scan(&a.Id, &a.Balance, &a.Frozen, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (r accountPostgresRepo) FindAll(c context.Context) ([]model.Account, error) {
	var accounts []model.Account
	rows, err := r.db.Query(c, fmt.Sprintf(`
		select id, balance, frozen, created_at, updated_at from %s
	`, model.AccountsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var account model.Account
	for rows.Next() {
		if err := rows.Scan(&account.Id, &account.Balance, &account.Frozen, &account.CreatedAt, &account.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r accountPostgresRepo) UpdateOne(c context.Context, accountId uint, balanceChange, frozenChange float64) error {
	_, err := r.db.Exec(c, fmt.Sprintf(`
		update %s
		set balance = balance+$1, 
			frozen = frozen+$2, 
			updated_at = current_timestamp
		where id = $3
	`, model.AccountsTable), balanceChange, frozenChange, accountId)
	return err
}
