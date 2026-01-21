package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			email:   "first.last@example.co.uk",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name:    "invalid email no @",
			email:   "userexample.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email no domain",
			email:   "user@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email no local part",
			email:   "@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid UUID",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "empty UUID",
			id:      "",
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name:    "invalid UUID format",
			id:      "not-a-uuid",
			wantErr: true,
			errMsg:  "invalid UUID format",
		},
		{
			name:    "invalid UUID missing dashes",
			id:      "550e8400e29b41d4a716446655440000",
			wantErr: true,
			errMsg:  "invalid UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid positive amount",
			amount:  12.50,
			wantErr: false,
		},
		{
			name:    "valid small amount",
			amount:  0.01,
			wantErr: false,
		},
		{
			name:    "zero amount",
			amount:  0,
			wantErr: true,
			errMsg:  "amount must be greater than 0",
		},
		{
			name:    "negative amount",
			amount:  -10.50,
			wantErr: true,
			errMsg:  "amount must be greater than 0",
		},
		{
			name:    "amount exceeds maximum",
			amount:  100000000.00,
			wantErr: true,
			errMsg:  "amount exceeds maximum value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCategoryName(t *testing.T) {
	tests := []struct {
		name    string
		nameStr string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid name",
			nameStr: "Food",
			wantErr: false,
		},
		{
			name:    "valid name with spaces",
			nameStr: "Food & Drinks",
			wantErr: false,
		},
		{
			name:    "empty name",
			nameStr: "",
			wantErr: true,
			errMsg:  "category name is required",
		},
		{
			name:    "whitespace only",
			nameStr: "   ",
			wantErr: true,
			errMsg:  "category name cannot be only whitespace",
		},
		{
			name:    "name too long",
			nameStr: string(make([]byte, 101)),
			wantErr: true,
			errMsg:  "category name cannot exceed 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategoryName(tt.nameStr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	t.Run("nil description", func(t *testing.T) {
		err := ValidateDescription(nil)
		assert.NoError(t, err)
	})

	t.Run("valid description", func(t *testing.T) {
		desc := "Lunch with friends"
		err := ValidateDescription(&desc)
		assert.NoError(t, err)
	})

	t.Run("empty description", func(t *testing.T) {
		desc := ""
		err := ValidateDescription(&desc)
		assert.NoError(t, err)
	})

	t.Run("description too long", func(t *testing.T) {
		desc := string(make([]byte, 1001))
		err := ValidateDescription(&desc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "description cannot exceed 1000 characters")
	})

	t.Run("description at max length", func(t *testing.T) {
		desc := string(make([]byte, 1000))
		err := ValidateDescription(&desc)
		assert.NoError(t, err)
	})
}

func TestValidateDateRange(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		from    *time.Time
		to      *time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name:    "both nil",
			from:    nil,
			to:      nil,
			wantErr: false,
		},
		{
			name:    "valid range",
			from:    &yesterday,
			to:      &tomorrow,
			wantErr: false,
		},
		{
			name:    "from after to",
			from:    &tomorrow,
			to:      &yesterday,
			wantErr: true,
			errMsg:  "'from' date must be before or equal to 'to' date",
		},
		{
			name:    "from nil, to valid",
			from:    nil,
			to:      &tomorrow,
			wantErr: false,
		},
		{
			name:    "from valid, to nil",
			from:    &yesterday,
			to:      nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
