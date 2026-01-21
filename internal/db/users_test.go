package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/models"
	"fintrack-go/tests/dbtestutil"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) CreateUser(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestCreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		email := "test-user@example.com"

		user, err := db.CreateUser(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, email, user.Email)
		assert.False(t, user.CreatedAt.IsZero())

		dbtestutil.AssertRowExists(t, pool, 
			"SELECT 1 FROM users WHERE id = $1", user.ID)
	})

	t.Run("duplicate email", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		email := "duplicate@example.com"
		_, err := db.CreateUser(ctx, email)
		require.NoError(t, err)

		_, err = db.CreateUser(ctx, email)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDuplicateEmail)
	})

	t.Run("empty email", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		_, err := db.CreateUser(ctx, "")
		require.Error(t, err)
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		created, err := db.CreateUser(ctx, "find-by-id@example.com")
		require.NoError(t, err)

		found, err := db.GetUserByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Email, found.Email)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		_, err := db.GetUserByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		email := "find-by-email@example.com"
		created, err := db.CreateUser(ctx, email)
		require.NoError(t, err)

		found, err := db.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		_, err := db.GetUserByEmail(ctx, "non-existent@example.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUserSQLInjection(t *testing.T) {
	t.Run("email injection", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		maliciousEmail := "'; DROP TABLE users; --"

		user, err := db.CreateUser(ctx, maliciousEmail)
		require.NoError(t, err)
		require.NotNil(t, user)

		dbtestutil.AssertRowCount(t, pool, 1, "SELECT COUNT(*) FROM users")
	})
}
