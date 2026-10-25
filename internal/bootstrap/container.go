// Package bootstrap собирает зависимости приложения через DI-контейнер.
package bootstrap

import (
	"fmt"
	"io"

	"go.uber.org/dig"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

// NewContainer регистрирует техосмотр, хранилище и сценарии приложения.
// Каждый контейнер создаёт собственный баланс. Подмены зависимостей
// через Decorate выполняются до первого Invoke.
func NewContainer() (*dig.Container, error) {
	container := dig.New()
	if err := container.Provide(func() application.Inspector {
		return domain.ServiceCenter{}
	}); err != nil {
		return nil, fmt.Errorf("регистрация техосмотра: %w", err)
	}
	if err := container.Provide(func() application.Repository {
		return &memory.FleetRepository{}
	}); err != nil {
		return nil, fmt.Errorf("регистрация хранилища: %w", err)
	}
	if err := container.Provide(application.NewFleet); err != nil {
		return nil, fmt.Errorf("регистрация сценариев приложения: %w", err)
	}
	return container, nil
}

// Run собирает приложение и запускает консольную сессию.
func Run(input io.Reader, output io.Writer) error {
	container, err := NewContainer()
	if err != nil {
		return err
	}
	if err := container.Invoke(func(fleet *application.Fleet) error {
		return cli.Run(fleet, input, output)
	}); err != nil {
		return fmt.Errorf("запуск приложения: %w", err)
	}
	return nil
}
