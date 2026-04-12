package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// handleListPosts → GET /api/posts
// Devuelve todos los posts ordenados por fecha descendente,
// cada uno con su lista de categorías.
func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := s.queries.ListPosts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("ListPosts failed")
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

// handleGetPost → GET /api/posts/{slug}
// Devuelve el post completo: metadata + categorías + bloques de contenido.
func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	post, err := s.queries.GetPostBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		log.Error().Err(err).Str("slug", slug).Msg("GetPostBySlug failed")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	cats, err := s.queries.GetCategoriesByPostID(ctx, post.ID)
	if err != nil {
		log.Error().Err(err).Str("post_id", post.ID.String()).Msg("GetCategoriesByPostID failed")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	blocks, err := s.queries.GetContentBlocksByPostID(ctx, post.ID)
	if err != nil {
		log.Error().Err(err).Str("post_id", post.ID.String()).Msg("GetContentBlocksByPostID failed")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, PostDetail{
		PostSummary: buildPostSummary(post, cats),
		Content:     buildContentBlocks(blocks),
	})
}
