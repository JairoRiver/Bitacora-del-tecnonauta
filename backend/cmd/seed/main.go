package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/config"
	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

// ── Estructuras para parsear posts.json ───────────────────────────────────

type postJSON struct {
	Title      string           `json:"Title"`
	Subtitle   string           `json:"Subtitle"`
	Date       string           `json:"Date"` // formato DD/MM/YYYY
	Categories []string         `json:"Categories"`
	HeroImage  string           `json:"HeroImage"`
	Slug       string           `json:"Slug"`
	Content    []blockJSON      `json:"Content"`
}

type blockJSON struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
	// conf está ausente en ImageBlock → llega como nil
	Conf json.RawMessage `json:"conf"`
}

// ── Main ──────────────────────────────────────────────────────────────────

func main() {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
	).With().Timestamp().Logger()

	inputFlag := flag.String("input", "../frontend/src/data/posts.json", "ruta al archivo posts.json")
	flag.Parse()

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

	posts, err := readPosts(*inputFlag)
	if err != nil {
		log.Fatal().Err(err).Str("file", *inputFlag).Msg("failed to read posts file")
	}

	log.Info().Int("count", len(posts)).Str("file", *inputFlag).Msg("posts loaded from file")

	queries := db.New(pool)

	imported, skipped := 0, 0
	for _, p := range posts {
		ok, err := seedPost(ctx, queries, p)
		if err != nil {
			log.Error().Err(err).Str("slug", p.Slug).Msg("failed to seed post")
			continue
		}
		if ok {
			imported++
		} else {
			skipped++
		}
	}

	log.Info().Int("imported", imported).Int("skipped", skipped).Msg("seed complete")
}

// ── Helpers ───────────────────────────────────────────────────────────────

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

func readPosts(path string) ([]postJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var posts []postJSON
	if err := json.Unmarshal(data, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// seedPost inserta un post junto con sus categorías y bloques de contenido.
// Es idempotente: si el slug ya existe, lo omite y devuelve (false, nil).
func seedPost(ctx context.Context, q db.Querier, p postJSON) (bool, error) {
	// Comprobar si ya existe
	_, err := q.GetPostBySlug(ctx, p.Slug)
	if err == nil {
		log.Debug().Str("slug", p.Slug).Msg("post already exists, skipping")
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("checking slug: %w", err)
	}

	date, err := time.Parse("02/01/2006", p.Date)
	if err != nil {
		return false, fmt.Errorf("parsing date %q: %w", p.Date, err)
	}

	post, err := q.CreatePost(ctx, db.CreatePostParams{
		Title:     p.Title,
		Subtitle:  p.Subtitle,
		Date:      date,
		HeroImage: p.HeroImage,
		Slug:      p.Slug,
	})
	if err != nil {
		return false, fmt.Errorf("creating post: %w", err)
	}

	// Categorías
	for _, name := range p.Categories {
		cat, err := q.GetOrCreateCategory(ctx, db.GetOrCreateCategoryParams{
			Name: name,
			Slug: slugify(name),
		})
		if err != nil {
			return false, fmt.Errorf("category %q: %w", name, err)
		}
		if err := q.AddPostCategory(ctx, db.AddPostCategoryParams{
			PostID:     post.ID,
			CategoryID: cat.ID,
		}); err != nil {
			return false, fmt.Errorf("linking category %q: %w", name, err)
		}
	}

	// Bloques de contenido
	for pos, block := range p.Content {
		conf := normalizeConf(block.Type, block.Conf)

		if _, err := q.CreateContentBlock(ctx, db.CreateContentBlockParams{
			PostID:   post.ID,
			Position: int32(pos),
			Type:     db.BlockType(block.Type),
			Value:    block.Value,
			Conf:     conf,
		}); err != nil {
			return false, fmt.Errorf("block %d: %w", pos, err)
		}
	}

	log.Info().Str("slug", p.Slug).Str("title", p.Title).Msg("post imported")
	return true, nil
}

// slugify replica la función parseRouteName del frontend TS.
func slugify(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// normalizeConf garantiza que los ImageBlock (sin campo conf en el JSON)
// almacenen JSON null en la DB en lugar de nil/vacío.
func normalizeConf(blockType string, conf json.RawMessage) json.RawMessage {
	if blockType == "i" {
		return json.RawMessage(`null`)
	}
	if len(conf) == 0 {
		return json.RawMessage(`null`)
	}
	return conf
}
