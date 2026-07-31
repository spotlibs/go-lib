package databases

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
)

// TransactionFunc is the type for a function that wraps logic inside a DB transaction.
type TransactionFunc func(fn func(tx orm.Query) error) error

// WithTransaction begins a database transaction, executes fn within it,
// and commits on success or rolls back on failure or panic.
func WithTransaction(fn func(tx orm.Query) error) error {
	return WithTransactionQuery(facades.Orm().Query(), fn)
}

// WithTransactionQuery is the testable core — accepts an orm.Query directly.
func WithTransactionQuery(q orm.Query, fn func(tx orm.Query) error) error {
	tx, err := q.Begin()
	if err != nil {
		return stderr.ErrInvRule("failed to begin transaction: " + err.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err = fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return stderr.ErrInvRule("failed to commit: " + err.Error())
	}
	return nil
}
