package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func ValidateUUID(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid UUID format")
	}
	// Ensure it's in the standard dashed format
	if !uuidRegex.MatchString(id) {
		return errors.New("invalid UUID format")
	}
	return nil
}

func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0, got %.2f", amount)
	}
	if amount > 99999999.99 {
		return fmt.Errorf("amount exceeds maximum value of 99999999.99, got %.2f", amount)
	}
	return nil
}

func ValidateCategoryName(name string) error {
	if name == "" {
		return errors.New("category name is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("category name cannot be only whitespace")
	}
	if len(name) > 100 {
		return fmt.Errorf("category name cannot exceed 100 characters, got %d", len(name))
	}
	return nil
}

func ValidateDescription(desc *string) error {
	if desc == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*desc)
	if len(trimmed) > 1000 {
		return fmt.Errorf("description cannot exceed 1000 characters, got %d", len(trimmed))
	}
	return nil
}

func ValidateDateRange(from, to *time.Time) error {
	if from == nil && to == nil {
		return nil
	}
	
	if from != nil && to != nil {
		if from.After(*to) {
			return errors.New("'from' date must be before or equal to 'to' date")
		}
	}
	
	return nil
}
