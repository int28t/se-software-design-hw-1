// Package domain содержит объекты и правила учёта транспорта.
package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidTransport означает некорректные данные транспорта.
var ErrInvalidTransport = errors.New("некорректные данные транспорта")

const (
	minBeginnerEase       = 1
	maxBeginnerEase       = 10
	beginnerEaseThreshold = 6
)

// Vehicle описывает транспорт, который можно проверить и поставить на баланс.
type Vehicle interface {
	InventoryItem
	IsServiceable() bool
	BeginnerEase() int
	SuitableForBeginner() bool
}

// ValidateVehicle проверяет наличие транспорта и корректность его данных.
func ValidateVehicle(vehicle Vehicle) error {
	if vehicle == nil {
		return fmt.Errorf("%w: транспорт не задан", ErrInvalidTransport)
	}
	return vehicle.Validate()
}

// Transport описывает транспорт до принятия решения о допуске.
// Поля закрыты, чтобы после техосмотра нельзя было изменить проверенные данные.
type Transport struct {
	inventoryNumber string
	name            string
	serviceable     bool
	beginnerEase    int
}

// NewTransport создаёт транспорт и проверяет обязательные поля.
func NewTransport(inventoryNumber, name string, serviceable bool, beginnerEase int) (Transport, error) {
	transport := Transport{
		inventoryNumber: strings.TrimSpace(inventoryNumber),
		name:            strings.TrimSpace(name),
		serviceable:     serviceable,
		beginnerEase:    beginnerEase,
	}
	if err := transport.Validate(); err != nil {
		return Transport{}, err
	}
	return transport, nil
}

// Validate проверяет данные, в том числе запрещает пустое значение Transport.
func (t Transport) Validate() error {
	if t.inventoryNumber == "" {
		return fmt.Errorf("%w: укажите инвентарный номер", ErrInvalidTransport)
	}
	if t.name == "" {
		return fmt.Errorf("%w: укажите название", ErrInvalidTransport)
	}
	if t.beginnerEase < minBeginnerEase || t.beginnerEase > maxBeginnerEase {
		return fmt.Errorf("%w: простота для новичка должна быть от %d до %d", ErrInvalidTransport, minBeginnerEase, maxBeginnerEase)
	}
	return nil
}

// InventoryNumber возвращает инвентарный номер транспорта.
func (t Transport) InventoryNumber() string {
	return t.inventoryNumber
}

// Name возвращает название транспорта.
func (t Transport) Name() string {
	return t.name
}

// IsServiceable сообщает о технической исправности транспорта.
func (t Transport) IsServiceable() bool {
	return t.serviceable
}

// BeginnerEase возвращает простоту управления по шкале от 1 до 10.
func (t Transport) BeginnerEase() int {
	return t.beginnerEase
}

// SuitableForBeginner проверяет правило простоты для новичка.
// Допуск по техническому состоянию проверяется отдельно при приёмке.
func (t Transport) SuitableForBeginner() bool {
	return t.beginnerEase >= beginnerEaseThreshold && t.beginnerEase <= maxBeginnerEase
}
