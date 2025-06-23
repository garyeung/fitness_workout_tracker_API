package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

func executeQueryRow(ctx context.Context, db *sql.DB, query string, args ...any) (*sql.Row, error) {

	stmr, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}

	defer stmr.Close()

	return stmr.QueryRowContext(ctx, args...), nil
}

func executeQuery(ctx context.Context, db *sql.DB, query string, args ...any) (*sql.Rows, error) {
	stmr, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}

	defer stmr.Close()

	return stmr.QueryContext(ctx, args...)
}

// For `Exec` operations.
func executeNonQuery(ctx context.Context, db *sql.DB, query string, args ...any) (sql.Result, error) {
	stmr, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}

	defer stmr.Close()

	result, err := stmr.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return result, nil

}

// For handling transactions.
func executeTransaction(ctx context.Context, db *sql.DB, txFunc func(context.Context, *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Rollback if any error occurs
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic after rollback
		}
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone { // Check if Rollback failed and the transaction wasn't already committed/rolled back
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()
	err = txFunc(ctx, tx) // execute the function with the transaction
	if err != nil {
		return fmt.Errorf("transaction function failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Successfully commit transcation")
	return nil
}
