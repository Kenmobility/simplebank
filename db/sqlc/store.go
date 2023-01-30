package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db Queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the successful transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

// TransferTx performs a money transfer from one account to the other.
// It creates a transfer record, add account entries, and update accounts' balance all within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		createTransferArgs := CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		}
		var err error

		txName := ctx.Value(txKey) //get the value of the key saved while calling the method using the passed context

		//we use q to access all the CRUD functions of Queries object
		fmt.Println(txName, " is creating transfer")
		result.Transfer, err = q.CreateTransfer(ctx, createTransferArgs) //try to insert into transfers table if any error, return from call back inorder for execTx to rollback the tx
		if err != nil {
			return err
		}

		createFromEntryArgs := CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		}

		fmt.Println(txName, " is creating entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, createFromEntryArgs) //try to insert into entries table for account making the transfer
		if err != nil {
			return err
		}

		createToEntryArgs := CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		}

		fmt.Println(txName, " is creating entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, createToEntryArgs) //try to insert into entries table for account receiving the transfer
		if err != nil {
			return err
		}

		//update accounts' balance

		//To update the FromAccount first, we get the account
		/*fmt.Println(txName, " is getting account1 for update")
		account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		//then, update the balance column by substracting the amount
		fmt.Println(txName, " is updating account1 balance")
		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID,
			Balance: account1.Balance - arg.Amount,
		})
		if err != nil {
			return err
		}
		*/

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}

			//to update the ToAccount, first, we get the account
			/*fmt.Println(txName, " is getting account2 for update")
			account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}

			//then, update the balance column by adding the amount
			fmt.Println(txName, " is updating account2 balance")
			result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID:      arg.ToAccountID,
				Balance: account2.Balance + arg.Amount,
			})
			if err != nil {
				return err
			}
			*/
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}

			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
