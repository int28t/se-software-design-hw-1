package cli

import (
	"errors"

	"escooter-sharing/internal/domain"
)

func (s *session) addItem(chargingStation bool) error {
	item, err := s.readItem(chargingStation)
	if errors.Is(err, domain.ErrInvalidItem) || errors.Is(err, errInvalidInput) {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	if err != nil {
		return err
	}
	if err := s.fleet.AddItem(item); err != nil {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	return s.printf("Вещь %q (%s) добавлена на баланс.\n", item.Name(), item.InventoryNumber())
}

func (s *session) readItem(chargingStation bool) (domain.InventoryItem, error) {
	number, name, err := s.readIdentity()
	if err != nil {
		return nil, err
	}
	if !chargingStation {
		return domain.NewHelmet(number, name)
	}
	if _, err := domain.NewItem(number, name); err != nil {
		return nil, err
	}
	energy, err := s.readEnergy()
	if err != nil {
		return nil, err
	}
	return domain.NewChargingStation(number, name, energy)
}

func (s *session) showInventory() error {
	entries, err := s.fleet.Inventory()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	if len(entries) == 0 {
		return s.printf("На балансе пока нет объектов.\n")
	}
	if err := s.printf("Инвентарная ведомость:\n"); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := s.printf("%s — %s\n", entry.InventoryNumber, entry.Name); err != nil {
			return err
		}
	}
	return nil
}
