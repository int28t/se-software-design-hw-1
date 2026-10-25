package application

import (
	"errors"
	"fmt"

	"escooter-sharing/internal/domain"
)

// ErrTransportRequiresInspection запрещает обход техосмотра через добавление вещи.
var ErrTransportRequiresInspection = errors.New("транспорт необходимо добавлять через техосмотр")

// InventoryEntry содержит строку инвентарной ведомости.
type InventoryEntry struct {
	InventoryNumber string
	Name            string
}

// AddItem ставит вещь на общий баланс без техосмотра транспорта.
func (f *Fleet) AddItem(item domain.InventoryItem) error {
	if err := domain.ValidateInventoryItem(item); err != nil {
		return fmt.Errorf("добавление вещи: %w", err)
	}
	if _, isTransport := item.(domain.Vehicle); isTransport {
		return ErrTransportRequiresInspection
	}
	if err := f.repository.Add(InventoryRecord{item: item}); err != nil {
		return fmt.Errorf("постановка вещи на баланс: %w", err)
	}
	return nil
}

// Inventory возвращает ведомость всех объектов на балансе по инвентарному номеру.
func (f *Fleet) Inventory() ([]InventoryEntry, error) {
	records, err := f.repository.List()
	if err != nil {
		return nil, fmt.Errorf("инвентарная ведомость: %w", err)
	}
	entries := make([]InventoryEntry, 0, len(records))
	for _, record := range records {
		item := record.Item()
		entries = append(entries, InventoryEntry{InventoryNumber: item.InventoryNumber(), Name: item.Name()})
	}
	return entries, nil
}
