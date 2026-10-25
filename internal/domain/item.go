package domain

import (
	"errors"
	"strings"
)

// ErrInvalidItem означает некорректные данные вещи на балансе.
var ErrInvalidItem = errors.New("некорректные данные вещи")

// Item содержит общие характеристики инвентаризируемых вещей.
type Item struct {
	inventoryNumber string
	name            string
}

// NewItem создаёт вещь с непустыми названием и инвентарным номером.
func NewItem(inventoryNumber, name string) (Item, error) {
	item := Item{inventoryNumber: strings.TrimSpace(inventoryNumber), name: strings.TrimSpace(name)}
	if err := item.Validate(); err != nil {
		return Item{}, err
	}
	return item, nil
}

// Validate проверяет обязательные поля вещи.
func (i Item) Validate() error {
	return validateInventoryFields(i.inventoryNumber, i.name, ErrInvalidItem)
}

// InventoryNumber возвращает инвентарный номер вещи.
func (i Item) InventoryNumber() string {
	return i.inventoryNumber
}

// Name возвращает название вещи.
func (i Item) Name() string {
	return i.name
}
