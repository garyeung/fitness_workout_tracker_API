package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DBVariables struct {
	Port     int
	Name     string
	Username string
	Password string
	Hostname string
}

func ConnectStr(dbVars DBVariables) string {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbVars.Hostname, dbVars.Port, dbVars.Username, dbVars.Password, dbVars.Name)
	return connStr

}

func NewPostgresDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Successfully connected to the database")

	return db, nil

}

func InitialTables(db *sql.DB, schemaFile string) error {

	tx, err := db.Begin() // start a transaction
	if err != nil {
		return fmt.Errorf("failed to begin trascation: %w", err)
	}

	defer tx.Rollback() // rollback if any error occurs

	sqlBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}
	sqlString := string(sqlBytes)

	_, err = tx.Exec(sqlString) // execute within the transaction
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit trascation: %w", err)
	}

	log.Println("Successfully executed database schema SQL")
	return nil

}
