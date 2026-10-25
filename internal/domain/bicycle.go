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
