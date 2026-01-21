package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/tests/dbtestutil"
)

func TestCreateTransaction(t *testing.T) {
	t.Run("success with category", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "txn-test@example.com")
		require.NoError(t, err)

		category, err := db.CreateCategory(ctx, user.ID, "Food")
		require.NoError(t, err)

		amount := 25.50
		desc := "Lunch"
		occurredAt := time.Now()

		transaction, err := db.CreateTransaction(ctx, user.ID, &category.ID, amount, &desc, occurredAt)
		require.NoError(t, err)
		require.NotNil(t, transaction)
		assert.NotEmpty(t, transaction.ID)
		assert.Equal(t, user.ID, transaction.UserID)
		assert.Equal(t, &category.ID, transaction.CategoryID)
		assert.Equal(t, amount, transaction.Amount)
		assert.Equal(t, &desc, transaction.Description)
		assert.False(t, transaction.CreatedAt.IsZero())

		dbtestutil.AssertRowExists(t, pool, 
			"SELECT 1 FROM transactions WHERE id = $1", transaction.ID)
	})

	t.Run("success without category", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "txn-no-cat@example.com")
		require.NoError(t, err)

		amount := 15.00
		transaction, err := db.CreateTransaction(ctx, user.ID, nil, amount, nil, time.Now())
		require.NoError(t, err)
		require.NotNil(t, transaction)
		assert.Nil(t, transaction.CategoryID)
	})

	t.Run("invalid amount", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "invalid-amt@example.com")
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, -10.0, nil, time.Now())
		require.Error(t, err)
	})

	t.Run("category not owned by user", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user1, err := db.CreateUser(ctx, "user1@example.com")
		require.NoError(t, err)

		user2, err := db.CreateUser(ctx, "user2@example.com")
		require.NoError(t, err)

		category, err := db.CreateCategory(ctx, user2.ID, "Private Category")
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user1.ID, &category.ID, 25.0, nil, time.Now())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to user")
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		nonExistentUserID := "550e8400-e29b-41d4-a716-446655440000"
		_, err := db.CreateTransaction(ctx, nonExistentUserID, nil, 10.0, nil, time.Now())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestListTransactions(t *testing.T) {
	t.Run("success with transactions", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "list-txn@example.com")
		require.NoError(t, err)

		category, err := db.CreateCategory(ctx, user.ID, "Food")
		require.NoError(t, err)

		now := time.Now()
		_, err = db.CreateTransaction(ctx, user.ID, &category.ID, 10.0, nil, now.Add(-3*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 20.0, nil, now.Add(-2*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, &category.ID, 30.0, nil, now.Add(-1*time.Hour))
		require.NoError(t, err)

		transactions, err := db.ListTransactions(ctx, user.ID, nil, nil)
		require.NoError(t, err)
		assert.Len(t, transactions, 3)
		assert.Equal(t, 30.0, transactions[0].Amount)
		assert.Equal(t, 20.0, transactions[1].Amount)
		assert.Equal(t, 10.0, transactions[2].Amount)
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "empty-list@example.com")
		require.NoError(t, err)

		transactions, err := db.ListTransactions(ctx, user.ID, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, transactions)
	})

	t.Run("date range filter", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "date-filter@example.com")
		require.NoError(t, err)

		now := time.Now()
		startDate := now.Add(-48 * time.Hour)
		endDate := now.Add(-24 * time.Hour)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 10.0, nil, now.Add(-72*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 20.0, nil, now.Add(-36*time.Hour))
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user.ID, nil, 30.0, nil, now.Add(-12*time.Hour))
		require.NoError(t, err)

		transactions, err := db.ListTransactions(ctx, user.ID, &startDate, &endDate)
		require.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.Equal(t, 20.0, transactions[0].Amount)
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

		_, err = db.CreateTransaction(ctx, user1.ID, nil, 10.0, nil, time.Now())
		require.NoError(t, err)

		_, err = db.CreateTransaction(ctx, user2.ID, nil, 20.0, nil, time.Now())
		require.NoError(t, err)

		user1Txns, err := db.ListTransactions(ctx, user1.ID, nil, nil)
		require.NoError(t, err)
		assert.Len(t, user1Txns, 1)
		assert.Equal(t, 10.0, user1Txns[0].Amount)
	})
}

func TestGetTransactionByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "get-txn@example.com")
		require.NoError(t, err)

		created, err := db.CreateTransaction(ctx, user.ID, nil, 25.50, nil, time.Now())
		require.NoError(t, err)

		found, err := db.GetTransactionByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Amount, found.Amount)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		_, err := db.GetTransactionByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})
}

func TestTransactionSQLInjection(t *testing.T) {
	t.Run("description injection", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		user, err := db.CreateUser(ctx, "sql-inj@example.com")
		require.NoError(t, err)

		maliciousDesc := "'; DROP TABLE transactions; --"
		_, err = db.CreateTransaction(ctx, user.ID, nil, 10.0, &maliciousDesc, time.Now())
		require.NoError(t, err)

		dbtestutil.AssertRowCount(t, pool, 1, "SELECT COUNT(*) FROM transactions")
	})

	t.Run("user_id injection", func(t *testing.T) {
		t.Parallel()

		ctx := dbtestutil.CreateTestContext(t)
		pool := dbtestutil.SetupTestDB(t)
		defer dbtestutil.TeardownTestDB(t, pool)

		db := &DB{pool: pool}
		maliciousUserID := "'; SELECT * FROM users; --"

		transactions, err := db.ListTransactions(ctx, maliciousUserID, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, transactions)
	})
}
