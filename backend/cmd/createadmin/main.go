package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/config"
	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

func main() {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
	).With().Timestamp().Logger()

	username := flag.String("username", "admin", "admin username")
	password := flag.String("password", "", "admin password (required)")
	flag.Parse()

	if *password == "" {
		log.Fatal().Msg("--password is required")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx := context.Background()
	pool, err := connect(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	q := db.New(pool)

	// Skip if user already exists
	_, err = q.GetUserByUsername(ctx, *username)
	if err == nil {
		log.Warn().Str("username", *username).Msg("user already exists, skipping")
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Fatal().Err(err).Msg("failed to check existing user")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to hash password")
	}

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Username:     *username,
		PasswordHash: string(hash),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create user")
	}

	log.Info().Str("username", user.Username).Msg("admin user created")
}

func connect(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return pool, nil
}
