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
