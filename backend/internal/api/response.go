package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

// PostSummary es la forma que devuelve GET /api/posts y GET /api/categories/{slug}.
// Incluye las categorías del post como lista de nombres.
type PostSummary struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Subtitle   string    `json:"subtitle"`
	Date       string    `json:"date"` // formato YYYY-MM-DD
	HeroImage  string    `json:"hero_image"`
	Slug       string    `json:"slug"`
	Categories []string  `json:"categories"`
}

// ContentBlockResponse espeja la union type ContentBlock del TS.
// conf es *json.RawMessage para que sea omitido (omitempty) cuando el bloque
// es de tipo "i" (ImageBlock), que almacena JSON null en la DB.
type ContentBlockResponse struct {
	Type  db.BlockType     `json:"type"`
	Value json.RawMessage  `json:"value"`
	Conf  *json.RawMessage `json:"conf,omitempty"`
}

// PostDetail es la forma que devuelve GET /api/posts/{slug}.
type PostDetail struct {
	PostSummary
	Content []ContentBlockResponse `json:"content"`
}

// CategoryResponse es cada elemento que devuelve GET /api/categories.
type CategoryResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// buildPostSummary convierte un db.Post + sus categorías en un PostSummary.
func buildPostSummary(post db.Post, cats []db.Category) PostSummary {
	names := make([]string, len(cats))
	for i, c := range cats {
		names[i] = c.Name
	}
	return PostSummary{
		ID:         post.ID,
		Title:      post.Title,
		Subtitle:   post.Subtitle,
		Date:       post.Date.Format("2006-01-02"),
		HeroImage:  post.HeroImage,
		Slug:       post.Slug,
		Categories: names,
	}
}

// buildContentBlocks convierte los db.ContentBlock en la respuesta JSON.
// Los bloques de tipo "i" (ImageBlock) tienen conf=null en la DB;
// se devuelven sin el campo conf para respetar la interface TS.
func buildContentBlocks(blocks []db.ContentBlock) []ContentBlockResponse {
	result := make([]ContentBlockResponse, len(blocks))
	for i, b := range blocks {
		var conf *json.RawMessage
		if !isJSONNull(b.Conf) {
			conf = &b.Conf
		}
		result[i] = ContentBlockResponse{
			Type:  b.Type,
			Value: b.Value,
			Conf:  conf,
		}
	}
	return result
}

// isJSONNull devuelve true cuando el RawMessage es nil, vacío o la literal "null".
func isJSONNull(raw json.RawMessage) bool {
	return len(raw) == 0 || string(raw) == "null"
}

// writeJSON serializa v como JSON y escribe status + body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("failed to encode JSON response")
	}
}

// writeError escribe una respuesta de error en JSON.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
