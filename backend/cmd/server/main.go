package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/api"
	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/config"
	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	setupLogger(cfg.Server.LogLevel)

	log.Info().
		Str("host", cfg.Server.Host).
		Str("port", cfg.Server.Port).
		Str("log_level", cfg.Server.LogLevel).
		Msg("config loaded")

	pool, err := connectDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	sessionDur, err := time.ParseDuration(cfg.Admin.SessionDuration)
	if err != nil {
		log.Fatal().Err(err).Str("value", cfg.Admin.SessionDuration).Msg("invalid session_duration")
	}
	if cfg.Admin.AuthSecret == "" {
		log.Fatal().Msg("AUTH_SECRET is required")
	}

	queries := db.New(pool)
	server := api.NewServer(queries, cfg.Admin.AuthSecret, sessionDur)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info().Str("addr", addr).Msg("server starting")

	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}

func connectDB(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("name", cfg.Name).
		Msg("database connected")

	return pool, nil
}

func setupLogger(level string) {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Warn().Str("level", level).Msg("unknown log level, defaulting to info")
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
}
