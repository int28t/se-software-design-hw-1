package cli

import (
	"errors"
	"fmt"

	"escooter-sharing/internal/domain"
)

type transportKind uint8

const (
	bicycleKind transportKind = iota
	scooterKind
	electricBicycleKind
)

func (s *session) admit(kind transportKind) error {
	transport, err := s.readTransport(kind)
	if errors.Is(err, domain.ErrInvalidTransport) || errors.Is(err, errInvalidInput) {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	if err != nil {
		return err
	}
	admission, err := s.fleet.Admit(transport)
	if err != nil {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	if admission.Approved() {
		return s.printf("Транспорт %q (%s) допущен в активный парк.\n", transport.Name(), transport.InventoryNumber())
	}
	return s.printf("Транспорт %q (%s) отклонён техосмотром; оставлен на балансе. В прокат не допущен.\n", transport.Name(), transport.InventoryNumber())
}

func (s *session) readTransport(kind transportKind) (domain.Vehicle, error) {
	number, name, err := s.readIdentity()
	if err != nil {
		return nil, err
	}
	serviceable, err := s.readCondition()
	if err != nil {
		return nil, err
	}
	beginnerEase, err := s.readBeginnerEase()
	if err != nil {
		return nil, err
	}
	if kind == bicycleKind {
		return domain.NewBicycle(number, name, serviceable, beginnerEase)
	}
	if _, err := domain.NewTransport(number, name, serviceable, beginnerEase); err != nil {
		return nil, err
	}
	dailyEnergyKWh, err := s.readEnergy()
	if err != nil {
		return nil, err
	}
	switch kind {
	case scooterKind:
		return domain.NewScooter(number, name, serviceable, beginnerEase, dailyEnergyKWh)
	case electricBicycleKind:
		return domain.NewElectricBicycle(number, name, serviceable, beginnerEase, dailyEnergyKWh)
	default:
		return nil, fmt.Errorf("%w: неизвестный вид транспорта", errInvalidInput)
	}
}

func (s *session) showActive() error {
	transports, err := s.fleet.ActiveTransports()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	if len(transports) == 0 {
		return s.printf("Активный парк пуст.\n")
	}
	if err := s.printf("Допущенный транспорт:\n"); err != nil {
		return err
	}
	for _, transport := range transports {
		if err := s.printf("%s — %s\n", transport.InventoryNumber(), transport.Name()); err != nil {
			return err
		}
	}
	return nil
}
