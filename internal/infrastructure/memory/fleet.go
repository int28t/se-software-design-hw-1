// Package memory хранит баланс в памяти одной сессии.
package memory

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

// FleetRepository хранит транспорт и вещи по общему инвентарному номеру.
// Используется последовательно, без конкурентного доступа.
// Нулевое значение готово к использованию.
type FleetRepository struct {
	admissions map[string]application.InventoryRecord
}

// Add добавляет запись, не перезаписывая уже занятый инвентарный номер.
func (r *FleetRepository) Add(admission application.InventoryRecord) error {
	if err := domain.ValidateInventoryItem(admission.Item()); err != nil {
		return fmt.Errorf("запись на баланс: %w", err)
	}
	number := admission.Item().InventoryNumber()
	if _, exists := r.admissions[number]; exists {
		return fmt.Errorf("%w: %s", application.ErrDuplicateInventoryNumber, number)
	}
	if r.admissions == nil {
		r.admissions = make(map[string]application.InventoryRecord)
	}
	r.admissions[number] = admission
	return nil
}

// List возвращает отдельный срез записей, упорядоченных по номеру.
func (r *FleetRepository) List() ([]application.InventoryRecord, error) {
	return slices.SortedFunc(maps.Values(r.admissions), func(a, b application.InventoryRecord) int {
		return cmp.Compare(a.Item().InventoryNumber(), b.Item().InventoryNumber())
	}), nil
}
