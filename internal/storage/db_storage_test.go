package storage

import (
	"context"
	"testing"
)

// Переопределяем функцию retryOperation для тестов
func init() {
	// Заменяем функцию retryOperation на более простую для тестов
	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}
}

// TestDBStorage_RetryOperationOverride проверяет, что мы можем переопределить функцию retryOperation
func TestDBStorage_RetryOperationOverride(t *testing.T) {
	// Сохраняем оригинальную функцию для восстановления после теста
	originalRetryOperation := retryOperation
	defer func() {
		retryOperation = originalRetryOperation
	}()

	// Счетчик вызовов
	callCount := 0

	// Переопределяем функцию retryOperation
	retryOperation = func(ctx context.Context, operation func() error) error {
		callCount++
		return operation()
	}

	// Проверяем, что наша функция вызывается
	ctx := context.Background()
	err := retryOperation(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("Expected 1 call, got: %d", callCount)
	}
}
