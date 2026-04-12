package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

func createPost(t *testing.T, q db.Querier, slug string) db.Post {
	t.Helper()
	post, err := q.CreatePost(context.Background(), db.CreatePostParams{
		Title:     "Test Post " + slug,
		Subtitle:  "Test Subtitle",
		Date:      time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		HeroImage: "https://example.com/hero.jpg",
		Slug:      slug,
	})
	require.NoError(t, err)
	return post
}

func TestCreatePost(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "create-post-test")

	assert.Equal(t, "Test Post create-post-test", post.Title)
	assert.Equal(t, "Test Subtitle", post.Subtitle)
	assert.Equal(t, "create-post-test", post.Slug)
	assert.Equal(t, "https://example.com/hero.jpg", post.HeroImage)
	assert.False(t, post.ID.String() == "00000000-0000-0000-0000-000000000000")
	assert.False(t, post.CreatedAt.IsZero())
}

func TestGetPostBySlug(t *testing.T) {
	q := newQuerier(t)
	created := createPost(t, q, "get-by-slug-test")

	fetched, err := q.GetPostBySlug(context.Background(), "get-by-slug-test")
	require.NoError(t, err)

	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Title, fetched.Title)
	assert.Equal(t, created.Slug, fetched.Slug)
}

func TestGetPostBySlug_NotFound(t *testing.T) {
	q := newQuerier(t)

	_, err := q.GetPostBySlug(context.Background(), "slug-que-no-existe")
	require.Error(t, err)
}

func TestListPosts_OrderedByDateDesc(t *testing.T) {
	q := newQuerier(t)

	older := db.CreatePostParams{
		Title: "Post Antiguo", Subtitle: "s", Slug: "post-antiguo",
		Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HeroImage: "",
	}
	newer := db.CreatePostParams{
		Title: "Post Reciente", Subtitle: "s", Slug: "post-reciente",
		Date: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), HeroImage: "",
	}

	_, err := q.CreatePost(context.Background(), older)
	require.NoError(t, err)
	_, err = q.CreatePost(context.Background(), newer)
	require.NoError(t, err)

	posts, err := q.ListPosts(context.Background())
	require.NoError(t, err)
	require.Len(t, posts, 2)

	assert.True(t, posts[0].Date.After(posts[1].Date), "los posts deben estar ordenados de más reciente a más antiguo")
}

func TestUpdatePost(t *testing.T) {
	q := newQuerier(t)
	original := createPost(t, q, "update-post-test")

	updated, err := q.UpdatePost(context.Background(), db.UpdatePostParams{
		ID:        original.ID,
		Title:     "Título Actualizado",
		Subtitle:  "Subtítulo Actualizado",
		Date:      time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		HeroImage: "https://example.com/new-hero.jpg",
		Slug:      "update-post-test",
	})
	require.NoError(t, err)

	assert.Equal(t, original.ID, updated.ID)
	assert.Equal(t, "Título Actualizado", updated.Title)
	assert.Equal(t, "Subtítulo Actualizado", updated.Subtitle)
	assert.True(t, updated.UpdatedAt.After(original.UpdatedAt) || updated.UpdatedAt.Equal(original.UpdatedAt))
}

func TestDeletePost(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "delete-post-test")

	err := q.DeletePost(context.Background(), post.ID)
	require.NoError(t, err)

	_, err = q.GetPostByID(context.Background(), post.ID)
	require.Error(t, err, "el post eliminado no debe encontrarse")
}

func TestListPostsByCategory(t *testing.T) {
	q := newQuerier(t)

	post := createPost(t, q, "post-con-categoria")
	cat, err := q.GetOrCreateCategory(context.Background(), db.GetOrCreateCategoryParams{
		Name: "Go", Slug: "go",
	})
	require.NoError(t, err)

	err = q.AddPostCategory(context.Background(), db.AddPostCategoryParams{
		PostID: post.ID, CategoryID: cat.ID,
	})
	require.NoError(t, err)

	// Post sin categoría (no debe aparecer en el resultado)
	createPost(t, q, "post-sin-categoria")

	posts, err := q.ListPostsByCategory(context.Background(), "go")
	require.NoError(t, err)
	require.Len(t, posts, 1)
	assert.Equal(t, post.ID, posts[0].ID)
}
