package validator

import (
	"testing"
	"time"
)

func BenchmarkValidateEmail_Valid(b *testing.B) {
	email := "test@example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateEmail(email)
	}
}

func BenchmarkValidateEmail_Invalid(b *testing.B) {
	email := "invalid-email"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateEmail(email)
	}
}

func BenchmarkValidateUUID_Valid(b *testing.B) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateUUID(id)
	}
}

func BenchmarkValidateUUID_Invalid(b *testing.B) {
	id := "not-a-uuid"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateUUID(id)
	}
}

func BenchmarkValidateAmount_Valid(b *testing.B) {
	amount := 10.50
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAmount(amount)
	}
}

func BenchmarkValidateAmount_Invalid(b *testing.B) {
	amount := -10.50
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAmount(amount)
	}
}

func BenchmarkValidateCategoryName_Valid(b *testing.B) {
	name := "Food & Drinks"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateCategoryName(name)
	}
}

func BenchmarkValidateCategoryName_Invalid(b *testing.B) {
	name := ""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateCategoryName(name)
	}
}

func BenchmarkValidateDescription_Valid(b *testing.B) {
	desc := "Test transaction description"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDescription(&desc)
	}
}

func BenchmarkValidateDescription_Long(b *testing.B) {
	desc := "A" + string(make([]byte, 999))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDescription(&desc)
	}
}

func BenchmarkValidateDateRange_Valid(b *testing.B) {
	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDateRange(&from, &to)
	}
}

func BenchmarkValidateDateRange_Invalid(b *testing.B) {
	now := time.Now()
	from := now
	to := now.Add(-24 * time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDateRange(&from, &to)
	}
}
