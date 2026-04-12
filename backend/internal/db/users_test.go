package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

func TestCreateUser(t *testing.T) {
	q := newQuerier(t)

	user, err := q.CreateUser(context.Background(), db.CreateUserParams{
		Username:     "admin",
		PasswordHash: "$2a$10$hashedpassword",
	})
	require.NoError(t, err)

	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "$2a$10$hashedpassword", user.PasswordHash)
	assert.False(t, user.ID.String() == "00000000-0000-0000-0000-000000000000")
	assert.False(t, user.CreatedAt.IsZero())
}

func TestGetUserByUsername(t *testing.T) {
	q := newQuerier(t)

	_, err := q.CreateUser(context.Background(), db.CreateUserParams{
		Username:     "jairo",
		PasswordHash: "$2a$10$hashedpassword",
	})
	require.NoError(t, err)

	user, err := q.GetUserByUsername(context.Background(), "jairo")
	require.NoError(t, err)

	assert.Equal(t, "jairo", user.Username)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	q := newQuerier(t)

	_, err := q.GetUserByUsername(context.Background(), "usuario-inexistente")
	require.Error(t, err)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	q := newQuerier(t)

	_, err := q.CreateUser(context.Background(), db.CreateUserParams{
		Username:     "duplicado",
		PasswordHash: "hash1",
	})
	require.NoError(t, err)

	_, err = q.CreateUser(context.Background(), db.CreateUserParams{
		Username:     "duplicado",
		PasswordHash: "hash2",
	})
	require.Error(t, err, "username duplicado debe violar la constraint UNIQUE")
}
