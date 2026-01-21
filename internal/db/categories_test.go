package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/tests/dbtestutil"
)

func TestCreateCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "category-test@example.com")
		require.NoError(t, err)

		category, err := db.CreateCategory(ctx, user.ID, "Food")
		require.NoError(t, err)
		require.NotNil(t, category)
		assert.NotEmpty(t, category.ID)
		assert.Equal(t, user.ID, category.UserID)
		assert.Equal(t, "Food", category.Name)
		assert.False(t, category.CreatedAt.IsZero())

		dbtestutil.AssertRowExists(t, pool, 
			"SELECT 1 FROM categories WHERE id = $1", category.ID)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		nonExistentUserID := "550e8400-e29b-41d4-a716-446655440000"

		_, err := db.CreateCategory(ctx, nonExistentUserID, "Food")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("duplicate category name", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "dup-cat@example.com")
		require.NoError(t, err)

		name := "Food"
		_, err = db.CreateCategory(ctx, user.ID, name)
		require.NoError(t, err)

		_, err = db.CreateCategory(ctx, user.ID, name)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDuplicateCategory)
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "empty-name@example.com")
		require.NoError(t, err)

		_, err = db.CreateCategory(ctx, user.ID, "")
		require.Error(t, err)
	})

	t.Run("name too long", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "long-name@example.com")
		require.NoError(t, err)

		longName := "Very long category name that exceeds the maximum allowed length of 100 characters"
		_, err = db.CreateCategory(ctx, user.ID, longName)
		require.Error(t, err)
	})
}

func TestListCategories(t *testing.T) {
	t.Run("success with categories", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "list-test@example.com")
		require.NoError(t, err)

		cat1, err := db.CreateCategory(ctx, user.ID, "Food")
		require.NoError(t, err)

		cat2, err := db.CreateCategory(ctx, user.ID, "Transport")
		require.NoError(t, err)

		categories, err := db.ListCategories(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, categories, 2)
		
		names := []string{categories[0].Name, categories[1].Name}
		assert.Contains(t, names, cat1.Name)
		assert.Contains(t, names, cat2.Name)
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "empty-list@example.com")
		require.NoError(t, err)

		categories, err := db.ListCategories(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, categories)
	})

	t.Run("user isolation", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user1, err := db.CreateUser(ctx, "user1@example.com")
		require.NoError(t, err)

		user2, err := db.CreateUser(ctx, "user2@example.com")
		require.NoError(t, err)

		cat1, err := db.CreateCategory(ctx, user1.ID, "User1 Category")
		require.NoError(t, err)

		_, err = db.CreateCategory(ctx, user2.ID, "User2 Category")
		require.NoError(t, err)

		user1Cats, err := db.ListCategories(ctx, user1.ID)
		require.NoError(t, err)
		assert.Len(t, user1Cats, 1)
		assert.Equal(t, cat1.ID, user1Cats[0].ID)
	})
}

func TestGetCategoryByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "get-cat@example.com")
		require.NoError(t, err)

		created, err := db.CreateCategory(ctx, user.ID, "Test Category")
		require.NoError(t, err)

		found, err := db.GetCategoryByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Name, found.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		_, err := db.GetCategoryByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

func TestCategorySQLInjection(t *testing.T) {
	t.Run("name injection", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "sql-injection@example.com")
		require.NoError(t, err)

		maliciousName := "'; DROP TABLE categories; --"
		_, err = db.CreateCategory(ctx, user.ID, maliciousName)
		require.NoError(t, err)

		dbtestutil.AssertRowCount(t, pool, 1, "SELECT COUNT(*) FROM categories")
	})
}
