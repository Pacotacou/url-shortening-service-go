package interfaces

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DBConfig holds database connection parameters
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// PostgresConnector handles database connections and operations
type PostgresConnector struct {
	db     *sql.DB
	config DBConfig
}

// NewPostgresConnector creates a new connector with the provided configuration
func NewPostgresConnector(config DBConfig) *PostgresConnector {
	return &PostgresConnector{
		config: config,
	}
}

// Connect establishes a connection to the database
func (p *PostgresConnector) Connect() error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.DBName,
		p.config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	log.Println("Successfully connected to PostgreSQL database")
	return nil
}

// Close terminates the database connection
func (p *PostgresConnector) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// GetDB returns the database instance
func (p *PostgresConnector) GetDB() *sql.DB {
	return p.db
}
