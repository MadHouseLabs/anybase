package postgres

import (
	"context"
	"database/sql"

	"github.com/madhouselabs/anybase/internal/database/types"
)

// PostgresTransaction wraps a SQL transaction to implement types.Transaction
type PostgresTransaction struct {
	tx      *sql.Tx
	adapter *PostgresAdapter
	ctx     context.Context
}

// Commit commits the transaction
func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// Collection returns a collection that uses this transaction
func (t *PostgresTransaction) Collection(name string) types.Collection {
	// For now, return the regular collection
	// TODO: Implement transaction-aware collection
	return t.adapter.Collection(name)
}

// Context returns the transaction context
func (t *PostgresTransaction) Context() context.Context {
	return t.ctx
}