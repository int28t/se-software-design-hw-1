// Package application реализует сценарии приёмки и допуска транспорта.
package application

import (
	"errors"
	"fmt"

	"escooter-sharing/internal/domain"
)

// ErrDuplicateInventoryNumber означает, что инвентарный номер уже занят.
var ErrDuplicateInventoryNumber = errors.New("инвентарный номер уже используется")

// Inspector проверяет транспорт перед его приёмкой.
// false без ошибки означает отказ по техническому состоянию.
type Inspector interface {
	Inspect(domain.Vehicle) (bool, error)
}

// Repository хранит результаты приёмки. Add не должен перезаписывать номер
// или изменять состояние при ошибке. List возвращает записи по номеру.
type Repository interface {
	Add(Admission) error
	List() ([]Admission, error)
}

// Admission хранит транспорт и неизменяемое решение техосмотра.
// Запись создаётся сценарием приёмки после вызова Inspector.
type Admission struct {
	transport domain.Vehicle
	approved  bool
}

// Transport возвращает транспорт, прошедший процедуру приёмки.
func (a Admission) Transport() domain.Vehicle {
	return a.transport
}

// Approved сообщает, разрешена ли выдача транспорта в прокат.
func (a Admission) Approved() bool {
	return a.approved
}

// Fleet выполняет сценарии учёта транспорта через внедрённые зависимости.
type Fleet struct {
	inspector  Inspector
	repository Repository
}

// NewFleet создаёт сервис с техосмотром и хранилищем.
// Обе зависимости должны быть заданы.
func NewFleet(inspector Inspector, repository Repository) *Fleet {
	return &Fleet{inspector: inspector, repository: repository}
}

// Admit проводит техосмотр и сохраняет его решение вместе с транспортом.
// Отклонённый транспорт остаётся на балансе, но не попадает в активный парк.
func (f *Fleet) Admit(transport domain.Vehicle) (Admission, error) {
	if err := domain.ValidateVehicle(transport); err != nil {
		return Admission{}, fmt.Errorf("приёмка: %w", err)
	}
	approved, err := f.inspector.Inspect(transport)
	if err != nil {
		return Admission{}, fmt.Errorf("техосмотр: %w", err)
	}
	admission := Admission{transport: transport, approved: approved}
	if err := f.repository.Add(admission); err != nil {
		return Admission{}, fmt.Errorf("постановка на баланс: %w", err)
	}
	return admission, nil
}

// ActiveTransports возвращает только допущенный в прокат транспорт.
func (f *Fleet) ActiveTransports() ([]domain.Vehicle, error) {
	admissions, err := f.repository.List()
	if err != nil {
		return nil, fmt.Errorf("активный парк: %w", err)
	}
	transports := make([]domain.Vehicle, 0, len(admissions))
	for _, admission := range admissions {
		if admission.Approved() {
			transports = append(transports, admission.Transport())
		}
	}
	return transports, nil
}
