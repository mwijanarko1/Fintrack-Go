package testutil

import (
	"time"

	"fintrack-go/internal/models"
)

type TestDataFactory struct {
	counter int
}

func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{counter: 0}
}

func (f *TestDataFactory) CreateUser(email string) *models.User {
	f.counter++
	return &models.User{
		ID:        generateTestUUID(f.counter, "user"),
		Email:     email,
		CreatedAt: time.Now(),
	}
}

func (f *TestDataFactory) CreateCategory(userID, name string) *models.Category {
	f.counter++
	return &models.Category{
		ID:        generateTestUUID(f.counter, "cat"),
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
	}
}

func (f *TestDataFactory) CreateTransaction(userID string, categoryID *string, amount float64, description *string) *models.Transaction {
	f.counter++
	return &models.Transaction{
		ID:          generateTestUUID(f.counter, "txn"),
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		OccurredAt:  time.Now(),
		CreatedAt:   time.Now(),
	}
}

func (f *TestDataFactory) CreateBatchTransactions(userID string, categoryID *string, count int) []models.Transaction {
	transactions := make([]models.Transaction, count)
	for i := 0; i < count; i++ {
		amount := float64(i + 1) * 10.0
		desc := "Transaction description"
		transactions[i] = *f.CreateTransaction(userID, categoryID, amount, &desc)
	}
	return transactions
}

func (f *TestDataFactory) CreateUserWithTransactions(transactionCount int) (*models.User, []models.Transaction) {
	userEmail := "factory-user@example.com"
	user := f.CreateUser(userEmail)

	var categoryID *string
	if transactionCount > 0 {
		category := f.CreateCategory(user.ID, "Factory Category")
		categoryID = &category.ID
	}

	transactions := f.CreateBatchTransactions(user.ID, categoryID, transactionCount)

	return user, transactions
}

func (f *TestDataFactory) CreateDateRangeTransaction(userID string, date time.Time, amount float64) *models.Transaction {
	f.counter++
	return &models.Transaction{
		ID:          generateTestUUID(f.counter, "txn"),
		UserID:      userID,
		CategoryID:  nil,
		Amount:      amount,
		Description:  strPtr("Date ranged transaction"),
		OccurredAt:  date,
		CreatedAt:   time.Now(),
	}
}

func (f *TestDataFactory) CreateTransactionsInRange(userID string, from, to time.Time, count int) []models.Transaction {
	transactions := make([]models.Transaction, count)
	duration := to.Sub(from)
	step := duration / time.Duration(count)

	for i := 0; i < count; i++ {
		occurredAt := from.Add(step * time.Duration(i))
		amount := float64(i+1) * 15.0
		transactions[i] = *f.CreateDateRangeTransaction(userID, occurredAt, amount)
	}

	return transactions
}

func generateTestUUID(counter int, prefix string) string {
	return generateTestUUIDHelper(prefix, counter)
}

func strPtr(s string) *string {
	return &s
}
