package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/models"
	"fintrack-go/tests/dbtestutil"
)

func TestGetSummary(t *testing.T) {
	t.Run("success with categories", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "summary@example.com")
		require.NoError(t, err)

		category1, err := db.CreateCategory(ctx, user.ID, "Food")
		require.NoError(t, err)

		category2, err := db.CreateCategory(ctx, user.ID, "Transport")
		require.NoError(t, err)

		now := time.Now()
		_, err = db.CreateTransaction(ctx, user.ID, &category1.ID, 10.0, nil, now.Add(-3*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, &category1.ID, 20.0, nil, now.Add(-2*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, &category2.ID, 30.0, nil, now.Add(-1*time.Hour))
		require.NoError(t, err)

		summary, err := db.GetSummary(ctx, user.ID, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Equal(t, user.ID, summary.UserID)
		assert.Len(t, summary.Categories, 2)

		foodTotal := findCategoryTotal(summary.Categories, "Food")
		assert.Equal(t, 30.0, foodTotal)

		transportTotal := findCategoryTotal(summary.Categories, "Transport")
		assert.Equal(t, 30.0, transportTotal)
	})

	t.Run("success with uncategorized", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "uncat@example.com")
		require.NoError(t, err)

		now := time.Now()
		_, err = db.CreateTransaction(ctx, user.ID, nil, 10.0, nil, now.Add(-2*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 20.0, nil, now.Add(-1*time.Hour))
		require.NoError(t, err)

		summary, err := db.GetSummary(ctx, user.ID, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Len(t, summary.Categories, 1)

		uncatTotal := findCategoryTotal(summary.Categories, "Uncategorized")
		assert.Equal(t, 30.0, uncatTotal)
	})

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "empty@example.com")
		require.NoError(t, err)

		summary, err := db.GetSummary(ctx, user.ID, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Len(t, summary.Categories, 0)
	})

	t.Run("date range filter", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "date-range@example.com")
		require.NoError(t, err)

		now := time.Now()
		_, err = db.CreateTransaction(ctx, user.ID, nil, 10.0, nil, now.Add(-72*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 20.0, nil, now.Add(-48*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 30.0, nil, now.Add(-24*time.Hour))
		require.NoError(t, err)

		startDate := now.Add(-50 * time.Hour)
		endDate := now.Add(-40 * time.Hour)

		summary, err := db.GetSummary(ctx, user.ID, &startDate, &endDate)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Equal(t, startDate, summary.From)
		assert.Equal(t, endDate, summary.To)
		assert.Len(t, summary.Categories, 1)
		assert.Equal(t, 20.0, summary.Categories[0].Total)
	})

	t.Run("date range boundaries", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "boundary@example.com")
		require.NoError(t, err)

		now := time.Now()
		startDate := now.Add(-48 * time.Hour)
		endDate := now.Add(-24 * time.Hour)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 10.0, nil, startDate)
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 20.0, nil, endDate)
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 30.0, nil, now.Add(-36*time.Hour))
		require.NoError(t, err)

		summary, err := db.GetSummary(ctx, user.ID, &startDate, &endDate)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Len(t, summary.Categories, 1)
		assert.Equal(t, 60.0, summary.Categories[0].Total)
	})

	t.Run("default date range (last 30 days)", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "default-range@example.com")
		require.NoError(t, err)

		now := time.Now()
		_, err = db.CreateTransaction(ctx, user.ID, nil, 10.0, nil, now.Add(-45*24*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 20.0, nil, now.Add(-20*24*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 30.0, nil, now.Add(-10*24*time.Hour))
		require.NoError(t, err)

		summary, err := db.GetSummary(ctx, user.ID, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Len(t, summary.Categories, 1)
		assert.Equal(t, 50.0, summary.Categories[0].Total)
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

		now := time.Now()
		_, err = db.CreateTransaction(ctx, user1.ID, nil, 100.0, nil, now)
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user2.ID, nil, 200.0, nil, now)
		require.NoError(t, err)

		summary1, err := db.GetSummary(ctx, user1.ID, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 100.0, summary1.Categories[0].Total)

		summary2, err := db.GetSummary(ctx, user2.ID, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 200.0, summary2.Categories[0].Total)
	})
}

func TestValidateCategoryOwnership(t *testing.T) {
	t.Run("valid ownership", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "ownership@example.com")
		require.NoError(t, err)

		category, err := db.CreateCategory(ctx, user.ID, "Test Category")
		require.NoError(t, err)

		err = db.ValidateCategoryOwnership(ctx, category.ID, user.ID)
		assert.NoError(t, err)
	})

	t.Run("invalid ownership", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user1, err := db.CreateUser(ctx, "owner1@example.com")
		require.NoError(t, err)

		user2, err := db.CreateUser(ctx, "owner2@example.com")
		require.NoError(t, err)

		category, err := db.CreateCategory(ctx, user1.ID, "Private Category")
		require.NoError(t, err)

		err = db.ValidateCategoryOwnership(ctx, category.ID, user2.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to user")
	})

	t.Run("category not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "not-found@example.com")
		require.NoError(t, err)

		nonExistentCatID := "550e8400-e29b-41d4-a716-4466554400000"

		err = db.ValidateCategoryOwnership(ctx, nonExistentCatID, user.ID)
		require.Error(t, err)
	})
}

func findCategoryTotal(categories []models.CategorySummary, name string) float64 {
	for _, cat := range categories {
		if cat.CategoryName == name {
			return cat.Total
		}
	}
	return 0.0
}
