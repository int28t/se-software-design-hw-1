// Package application реализует сценарии приёмки транспорта и учёта баланса.
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

// Repository хранит записи баланса. Add не должен перезаписывать номер
// или изменять состояние при ошибке. List возвращает записи по номеру.
type Repository interface {
	Add(InventoryRecord) error
	List() ([]InventoryRecord, error)
}

// InventoryRecord хранит объект на балансе и решение техосмотра транспорта.
// Вещи не проходят техосмотр и не имеют допуска в активный парк.
type InventoryRecord struct {
	item     domain.InventoryItem
	approved bool
}

// Item возвращает инвентаризируемый объект.
func (r InventoryRecord) Item() domain.InventoryItem {
	return r.item
}

// Approved сообщает, разрешена ли выдача транспорта в прокат.
// Для вещей всегда возвращает false.
func (r InventoryRecord) Approved() bool {
	return r.approved
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
func (f *Fleet) Admit(transport domain.Vehicle) (InventoryRecord, error) {
	if err := domain.ValidateVehicle(transport); err != nil {
		return InventoryRecord{}, fmt.Errorf("приёмка: %w", err)
	}
	approved, err := f.inspector.Inspect(transport)
	if err != nil {
		return InventoryRecord{}, fmt.Errorf("техосмотр: %w", err)
	}
	admission := InventoryRecord{item: transport, approved: approved}
	if err := f.repository.Add(admission); err != nil {
		return InventoryRecord{}, fmt.Errorf("постановка на баланс: %w", err)
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
		if transport, ok := admission.Item().(domain.Vehicle); ok && admission.Approved() {
			transports = append(transports, transport)
		}
	}
	return transports, nil
}

// BeginnerTransports возвращает допущенную технику, подходящую новичкам.
func (f *Fleet) BeginnerTransports() ([]domain.Vehicle, error) {
	active, err := f.ActiveTransports()
	if err != nil {
		return nil, fmt.Errorf("техника для новичков: %w", err)
	}
	transports := make([]domain.Vehicle, 0, len(active))
	for _, transport := range active {
		if transport.SuitableForBeginner() {
			transports = append(transports, transport)
		}
	}
	return transports, nil
}
