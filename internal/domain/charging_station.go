package domain

import (
	"fmt"
	"math"
)

// ChargingStation описывает зарядную станцию на балансе.
type ChargingStation struct {
	Item
	dailyEnergyKWh float64
}

// NewChargingStation создаёт станцию с заданным суточным расходом энергии.
func NewChargingStation(inventoryNumber, name string, dailyEnergyKWh float64) (ChargingStation, error) {
	item, err := NewItem(inventoryNumber, name)
	if err != nil {
		return ChargingStation{}, err
	}
	station := ChargingStation{Item: item, dailyEnergyKWh: dailyEnergyKWh}
	if err := station.Validate(); err != nil {
		return ChargingStation{}, err
	}
	return station, nil
}

// Validate проверяет данные станции и конечный неотрицательный расход энергии.
func (s ChargingStation) Validate() error {
	if err := s.Item.Validate(); err != nil {
		return err
	}
	if s.dailyEnergyKWh < 0 || math.IsNaN(s.dailyEnergyKWh) || math.IsInf(s.dailyEnergyKWh, 0) {
		return fmt.Errorf("%w: расход энергии должен быть конечным числом не меньше нуля", ErrInvalidItem)
	}
	return nil
}

// DailyEnergyKWh возвращает заданный суточный расход станции в кВт·ч.
func (s ChargingStation) DailyEnergyKWh() float64 {
	return s.dailyEnergyKWh
}
