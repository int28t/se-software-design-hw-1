package domain

import "fmt"

// InventoryItem описывает объект на балансе с названием и инвентарным номером.
// Сотрудники этой роли не реализуют.
// После постановки на баланс его характеристики не должны изменяться.
type InventoryItem interface {
	InventoryNumber() string
	Name() string
	Validate() error
}

// ValidateInventoryItem проверяет наличие и корректность объекта на балансе.
func ValidateInventoryItem(item InventoryItem) error {
	if item == nil {
		return fmt.Errorf("%w: объект не задан", ErrInvalidItem)
	}
	return item.Validate()
}

func validateInventoryFields(number, name string, invalid error) error {
	if number == "" {
		return fmt.Errorf("%w: укажите инвентарный номер", invalid)
	}
	if name == "" {
		return fmt.Errorf("%w: укажите название", invalid)
	}
	return nil
}
