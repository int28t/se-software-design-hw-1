package domain

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
	if err := validateDailyEnergyKWh(dailyEnergyKWh, ErrInvalidItem); err != nil {
		return ChargingStation{}, err
	}
	return ChargingStation{Item: item, dailyEnergyKWh: dailyEnergyKWh}, nil
}

// Validate проверяет данные станции и конечный неотрицательный расход энергии.
func (s ChargingStation) Validate() error {
	if err := s.Item.Validate(); err != nil {
		return err
	}
	return validateDailyEnergyKWh(s.dailyEnergyKWh, ErrInvalidItem)
}

// DailyEnergyKWh возвращает заданный суточный расход станции в кВт·ч.
func (s ChargingStation) DailyEnergyKWh() float64 {
	return s.dailyEnergyKWh
}
