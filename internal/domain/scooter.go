package domain

// Scooter описывает электросамокат с заданным суточным расходом энергии.
type Scooter struct {
	ElectricTransport
}

// NewScooter создаёт электросамокат с проверенными характеристиками транспорта.
func NewScooter(inventoryNumber, name string, serviceable bool, beginnerEase int, dailyEnergyKWh float64) (Scooter, error) {
	transport, err := NewElectricTransport(inventoryNumber, name, serviceable, beginnerEase, dailyEnergyKWh)
	if err != nil {
		return Scooter{}, err
	}
	return Scooter{ElectricTransport: transport}, nil
}
