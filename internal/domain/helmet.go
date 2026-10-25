package domain

// Helmet описывает шлем, учитываемый на балансе.
type Helmet struct {
	Item
}

// NewHelmet создаёт шлем с названием и инвентарным номером.
func NewHelmet(inventoryNumber, name string) (Helmet, error) {
	item, err := NewItem(inventoryNumber, name)
	if err != nil {
		return Helmet{}, err
	}
	return Helmet{Item: item}, nil
}
