package chats_db

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/logger"
	"context"
	"fmt"

	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ChatDB struct to hold the database connection pool
type ChatDB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new database connection pool
// and returns a pointer to the ChatDB struct
func NewDB() (*ChatDB, error) {
	dbConfig := config.AppConfig.ChatDBConfig
	dbUser := dbConfig.User
	dbHost := dbConfig.Host
	dbPort := dbConfig.Port
	dbName := dbConfig.Name
	dbPass := dbConfig.Password
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatal("error while parsing config")
		return nil, fmt.Errorf("error while parsing config: %w", err)
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		pool.Close()
		log.Fatal("error while connecting new with config")
		return nil, fmt.Errorf("error while connecting new with config: %v", err)
	}

	// Create tables if not exists
	// how to make pair of title and type unique

	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS anna_chats(
		id SERIAL PRIMARY KEY,
		user_name TEXT,
		query TEXT,
		response TEXT,
		feedback INT,
		input_tokens INT,
		output_token INT,
		model TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	)`)

	if err != nil {
		logger.Logger.Errorf("error creating chats table: %v", err)
		return nil, fmt.Errorf("error creating chats table: %v", err)
	}
	// add indexing on user_name
	_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_user_name ON anna_chats(user_name)`)
	if err != nil {
		logger.Logger.Errorf("error creating index on user_name: %v", err)
		return nil, fmt.Errorf("error creating index on user_name: %v", err)
	}
	logger.Logger.Info("Connected to Chats DB successfully")
	return &ChatDB{Pool: pool}, nil
}

// Close closes the database connection pool

func (db *ChatDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
