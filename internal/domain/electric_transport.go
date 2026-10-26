package domain

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
	if err := validateDailyEnergyKWh(dailyEnergyKWh, ErrInvalidTransport); err != nil {
		return ElectricTransport{}, err
	}
	return ElectricTransport{Transport: transport, dailyEnergyKWh: dailyEnergyKWh}, nil
}

// Validate проверяет данные транспорта и суточный расход энергии.
func (t ElectricTransport) Validate() error {
	if err := t.Transport.Validate(); err != nil {
		return err
	}
	return validateDailyEnergyKWh(t.dailyEnergyKWh, ErrInvalidTransport)
}

// DailyEnergyKWh возвращает заданный суточный расход в кВт·ч.
func (t ElectricTransport) DailyEnergyKWh() float64 {
	return t.dailyEnergyKWh
}
