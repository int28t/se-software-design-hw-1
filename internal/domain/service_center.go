package domain

// ServiceCenter принимает решение о допуске по технической исправности.
type ServiceCenter struct{}

// Inspect допускает исправный транспорт и отклоняет неисправный.
// Некорректные данные являются ошибкой, а не отказом в техосмотре.
func (ServiceCenter) Inspect(transport Vehicle) (bool, error) {
	if err := ValidateVehicle(transport); err != nil {
		return false, err
	}
	return transport.IsServiceable(), nil
}
