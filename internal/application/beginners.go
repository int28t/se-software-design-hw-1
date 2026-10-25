package application

import (
	"fmt"

	"escooter-sharing/internal/domain"
)

// BeginnerTransports возвращает допущенную технику, подходящую новичкам.
func (f *Fleet) BeginnerTransports() ([]domain.Vehicle, error) {
	active, err := f.ActiveTransports()
	if err != nil {
		return nil, fmt.Errorf("техника для новичков: %w", err)
	}
	transports := make([]domain.Vehicle, 0, len(active))
	for _, transport := range active {
		if transport.SuitableForBeginner() {
			transports = append(transports, transport)
		}
	}
	return transports, nil
}
