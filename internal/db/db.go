package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Import the PostgreSQL driver as a side effect.
	// The underscore import registers the "postgres" driver with database/sql
	// without directly using any of its exported symbols.
	_ "github.com/lib/pq"
)

// DB is the package-level database connection pool.
// It is initialized by InitDB() and used directly by other packages via db.DB.
var DB *sql.DB

// InitDB creates and verifies a PostgreSQL connection using environment variables.
// It reads DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, and DB_NAME from the environment.
// Returns the connection pool and also stores it in the package-level DB variable
// so other packages can access it without passing it as a parameter.
// Calls log.Fatal if the connection cannot be opened or pinged — the app cannot
// run without a database connection.
func InitDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Build the PostgreSQL connection string from environment variables.
	// sslmode=disable is appropriate for local/Docker development.
	// In production, this should be changed to sslmode=require.
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	log.Println("📦 Connecting to DB at:", host)

	var err error

	// sql.Open validates the connection string format but does NOT
	// establish a real network connection yet.
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Failed to open DB:", err)
	}

	// DB.Ping() establishes the actual connection and verifies credentials.
	// This is where connection failures (wrong password, DB not running) are caught.
	err = DB.Ping()
	if err != nil {
		log.Fatal("❌ Failed to ping DB:", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	return DB
}
