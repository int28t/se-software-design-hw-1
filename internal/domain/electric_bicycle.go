package domain

// ElectricBicycle описывает электровелосипед с заданным суточным расходом энергии.
type ElectricBicycle struct {
	ElectricTransport
}

// NewElectricBicycle создаёт электровелосипед с проверенными характеристиками транспорта.
func NewElectricBicycle(inventoryNumber, name string, serviceable bool, beginnerEase int, dailyEnergyKWh float64) (ElectricBicycle, error) {
	transport, err := NewElectricTransport(inventoryNumber, name, serviceable, beginnerEase, dailyEnergyKWh)
	if err != nil {
		return ElectricBicycle{}, err
	}
	return ElectricBicycle{ElectricTransport: transport}, nil
}
