package domain

// Bicycle описывает велосипед без электропривода.
type Bicycle struct {
	Transport
}

// NewBicycle создаёт велосипед с проверенными характеристиками транспорта.
func NewBicycle(inventoryNumber, name string, serviceable bool, beginnerEase int) (Bicycle, error) {
	transport, err := NewTransport(inventoryNumber, name, serviceable, beginnerEase)
	if err != nil {
		return Bicycle{}, err
	}
	return Bicycle{Transport: transport}, nil
}

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
