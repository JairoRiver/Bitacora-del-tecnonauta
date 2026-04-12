package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

func createCategory(t *testing.T, q db.Querier, name, slug string) db.Category {
	t.Helper()
	cat, err := q.GetOrCreateCategory(context.Background(), db.GetOrCreateCategoryParams{
		Name: name,
		Slug: slug,
	})
	require.NoError(t, err)
	return cat
}

func TestGetOrCreateCategory_Create(t *testing.T) {
	q := newQuerier(t)

	cat := createCategory(t, q, "Python", "python")

	assert.Equal(t, "Python", cat.Name)
	assert.Equal(t, "python", cat.Slug)
	assert.False(t, cat.ID.String() == "00000000-0000-0000-0000-000000000000")
}

func TestGetOrCreateCategory_Idempotent(t *testing.T) {
	q := newQuerier(t)

	first := createCategory(t, q, "Rust", "rust")
	second := createCategory(t, q, "Rust", "rust")

	assert.Equal(t, first.ID, second.ID, "llamar dos veces con el mismo nombre debe devolver el mismo registro")
}

func TestListCategories_SortedAlphabetically(t *testing.T) {
	q := newQuerier(t)

	createCategory(t, q, "TypeScript", "typescript")
	createCategory(t, q, "Go", "go-lang")
	createCategory(t, q, "DuckDB", "duckdb")

	categories, err := q.ListCategories(context.Background())
	require.NoError(t, err)
	require.Len(t, categories, 3)

	assert.Equal(t, "DuckDB", categories[0].Name)
	assert.Equal(t, "Go", categories[1].Name)
	assert.Equal(t, "TypeScript", categories[2].Name)
}

func TestGetCategoriesByPostID(t *testing.T) {
	q := newQuerier(t)

	post := createPost(t, q, "post-cats-test")
	catGo := createCategory(t, q, "Go", "go")
	catSQL := createCategory(t, q, "SQL", "sql")

	err := q.AddPostCategory(context.Background(), db.AddPostCategoryParams{PostID: post.ID, CategoryID: catGo.ID})
	require.NoError(t, err)
	err = q.AddPostCategory(context.Background(), db.AddPostCategoryParams{PostID: post.ID, CategoryID: catSQL.ID})
	require.NoError(t, err)

	cats, err := q.GetCategoriesByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	require.Len(t, cats, 2)

	names := []string{cats[0].Name, cats[1].Name}
	assert.Contains(t, names, "Go")
	assert.Contains(t, names, "SQL")
}

func TestGetCategoriesByPostID_Empty(t *testing.T) {
	q := newQuerier(t)

	post := createPost(t, q, "post-no-cats")

	cats, err := q.GetCategoriesByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	assert.Empty(t, cats)
}

func TestClearPostCategories(t *testing.T) {
	q := newQuerier(t)

	post := createPost(t, q, "post-clear-cats")
	cat := createCategory(t, q, "Datos", "datos")

	err := q.AddPostCategory(context.Background(), db.AddPostCategoryParams{PostID: post.ID, CategoryID: cat.ID})
	require.NoError(t, err)

	err = q.ClearPostCategories(context.Background(), post.ID)
	require.NoError(t, err)

	cats, err := q.GetCategoriesByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	assert.Empty(t, cats, "después de ClearPostCategories no debe haber categorías asociadas")
}

func TestAddPostCategory_Idempotent(t *testing.T) {
	q := newQuerier(t)

	post := createPost(t, q, "post-dup-cat")
	cat := createCategory(t, q, "Rust", "rust-dup")

	params := db.AddPostCategoryParams{PostID: post.ID, CategoryID: cat.ID}

	err := q.AddPostCategory(context.Background(), params)
	require.NoError(t, err)
	err = q.AddPostCategory(context.Background(), params)
	require.NoError(t, err, "ON CONFLICT DO NOTHING: insertar duplicado no debe fallar")

	cats, err := q.GetCategoriesByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	assert.Len(t, cats, 1)
}
