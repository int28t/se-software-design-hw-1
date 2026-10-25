// Package memory хранит результаты приёмки в памяти одной сессии.
package memory

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

// FleetRepository хранит записи приёмки по инвентарному номеру.
// Используется последовательно, без конкурентного доступа.
// Нулевое значение готово к использованию.
type FleetRepository struct {
	admissions map[string]application.Admission
}

// Add добавляет запись, не перезаписывая уже занятый инвентарный номер.
func (r *FleetRepository) Add(admission application.Admission) error {
	if err := domain.ValidateVehicle(admission.Transport()); err != nil {
		return fmt.Errorf("запись на баланс: %w", err)
	}
	number := admission.Transport().InventoryNumber()
	if _, exists := r.admissions[number]; exists {
		return fmt.Errorf("%w: %s", application.ErrDuplicateInventoryNumber, number)
	}
	if r.admissions == nil {
		r.admissions = make(map[string]application.Admission)
	}
	r.admissions[number] = admission
	return nil
}

// List возвращает отдельный срез записей, упорядоченных по номеру.
func (r *FleetRepository) List() ([]application.Admission, error) {
	return slices.SortedFunc(maps.Values(r.admissions), func(a, b application.Admission) int {
		return cmp.Compare(a.Transport().InventoryNumber(), b.Transport().InventoryNumber())
	}), nil
}
