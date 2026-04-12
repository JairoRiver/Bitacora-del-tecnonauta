package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// handleListCategories → GET /api/categories
// Devuelve todas las categorías ordenadas alfabéticamente.
func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cats, err := s.queries.ListCategories(ctx)
	if err != nil {
		log.Error().Err(err).Msg("ListCategories failed")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]CategoryResponse, len(cats))
	for i, c := range cats {
		result[i] = CategoryResponse{Name: c.Name, Slug: c.Slug}
	}

	writeJSON(w, http.StatusOK, result)
}

// handleListPostsByCategory → GET /api/categories/{slug}
// Devuelve los posts que pertenecen a la categoría indicada,
// ordenados por fecha descendente.
func (s *Server) handleListPostsByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	posts, err := s.queries.ListPostsByCategory(ctx, slug)
	if err != nil {
		log.Error().Err(err).Str("slug", slug).Msg("ListPostsByCategory failed")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]PostSummary, len(posts))
	for i, post := range posts {
		cats, err := s.queries.GetCategoriesByPostID(ctx, post.ID)
		if err != nil {
			log.Error().Err(err).Str("post_id", post.ID.String()).Msg("GetCategoriesByPostID failed")
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		result[i] = buildPostSummary(post, cats)
	}

	writeJSON(w, http.StatusOK, result)
}
