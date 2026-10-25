package domain

import (
	"fmt"
	"math"
)

// ElectricTransport дополняет транспорт суточным расходом электроэнергии.
type ElectricTransport struct {
	Transport
	dailyEnergyKWh float64
}

// NewElectricTransport создаёт электротранспорт с конечным неотрицательным расходом.
func NewElectricTransport(inventoryNumber, name string, serviceable bool, beginnerEase int, dailyEnergyKWh float64) (ElectricTransport, error) {
	transport, err := NewTransport(inventoryNumber, name, serviceable, beginnerEase)
	if err != nil {
		return ElectricTransport{}, err
	}
	electric := ElectricTransport{Transport: transport, dailyEnergyKWh: dailyEnergyKWh}
	if err := electric.Validate(); err != nil {
		return ElectricTransport{}, err
	}
	return electric, nil
}

// Validate проверяет данные транспорта и суточный расход энергии.
func (t ElectricTransport) Validate() error {
	if err := t.Transport.Validate(); err != nil {
		return err
	}
	if t.dailyEnergyKWh < 0 || math.IsNaN(t.dailyEnergyKWh) || math.IsInf(t.dailyEnergyKWh, 0) {
		return fmt.Errorf("%w: расход энергии должен быть конечным числом не меньше нуля", ErrInvalidTransport)
	}
	return nil
}

// DailyEnergyKWh возвращает заданный суточный расход в кВт·ч.
func (t ElectricTransport) DailyEnergyKWh() float64 {
	return t.dailyEnergyKWh
}
