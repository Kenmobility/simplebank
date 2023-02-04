package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kenmobility/simplebank/util"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashedPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.LastUpdatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)
	userGot, err := testQueries.GetUser(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, userGot)

	require.Equal(t, user.Username, userGot.Username)
	require.Equal(t, user.HashedPassword, userGot.HashedPassword)
	require.Equal(t, user.FullName, userGot.FullName)
	require.Equal(t, user.Email, userGot.Email)
	require.Equal(t, user.LastUpdatedAt, userGot.LastUpdatedAt)
	require.WithinDuration(t, user.CreatedAt, userGot.CreatedAt, time.Second)
	require.WithinDuration(t, user.LastUpdatedAt, userGot.LastUpdatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, userGot.PasswordChangedAt, time.Second)
}
